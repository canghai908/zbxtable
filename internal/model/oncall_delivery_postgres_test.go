package model

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

type onCallSMTPMessage struct{ from, to, raw string }

// A real loopback SMTP socket exercises the production mail library without
// consulting or contacting any developer SMTP account.
func onCallSMTPReceiver(t *testing.T, reject bool) (string, <-chan onCallSMTPMessage) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	received := make(chan onCallSMTPMessage, 1)
	done := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			done <- err
			return
		}
		defer conn.Close()
		if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
			done <- err
			return
		}
		wire := textproto.NewConn(conn)
		if err := wire.PrintfLine("220 localhost SMTP test"); err != nil {
			done <- err
			return
		}
		var message onCallSMTPMessage
		for {
			line, err := wire.ReadLine()
			if err == io.EOF {
				done <- nil
				return
			}
			if err != nil {
				done <- err
				return
			}
			var reply string
			switch {
			case strings.HasPrefix(line, "EHLO "):
				reply = "250-localhost\r\n250 AUTH CRAM-MD5"
			case strings.HasPrefix(line, "HELO "):
				reply = "250 localhost"
			case strings.HasPrefix(line, "AUTH CRAM-MD5"):
				challenge := []byte("<oncall-loopback-test@localhost>")
				if err := wire.PrintfLine("334 %s", base64.StdEncoding.EncodeToString(challenge)); err != nil {
					done <- err
					return
				}
				answer, err := wire.ReadLine()
				if err != nil {
					done <- err
					return
				}
				decoded, err := base64.StdEncoding.DecodeString(answer)
				mac := hmac.New(md5.New, []byte("loopback-only-secret"))
				_, _ = mac.Write(challenge)
				if err != nil || string(decoded) != fmt.Sprintf("sender@example.test %x", mac.Sum(nil)) {
					done <- fmt.Errorf("unexpected SMTP test credentials")
					return
				}
				reply = "235 authenticated"
			case strings.HasPrefix(line, "MAIL FROM:"):
				message.from = strings.TrimPrefix(line, "MAIL FROM:")
				reply = "250 sender accepted"
			case strings.HasPrefix(line, "RCPT TO:"):
				message.to = strings.TrimPrefix(line, "RCPT TO:")
				reply = "250 recipient accepted"
			case line == "DATA":
				if err := wire.PrintfLine("354 end with dot"); err != nil {
					done <- err
					return
				}
				body, err := io.ReadAll(wire.DotReader())
				if err != nil {
					done <- err
					return
				}
				message.raw = string(body)
				received <- message
				reply = "250 queued"
				if reject {
					reply = "550 test delivery rejected"
				}
			case line == "QUIT":
				done <- wire.PrintfLine("221 bye")
				return
			default:
				reply = "250 ok"
			}
			if err := wire.PrintfLine("%s", reply); err != nil {
				done <- err
				return
			}
		}
	}()
	t.Cleanup(func() {
		_ = listener.Close()
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("SMTP receiver: %v", err)
			}
		case <-time.After(6 * time.Second):
			t.Error("SMTP receiver did not stop")
		}
	})
	return listener.Addr().String(), received
}

func onCallMailBody(t *testing.T, header mail.Header, body io.Reader) string {
	t.Helper()
	media, params, err := mime.ParseMediaType(header.Get("Content-Type"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(media, "multipart/") {
		parts := multipart.NewReader(body, params["boundary"])
		var content strings.Builder
		for {
			part, err := parts.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			content.WriteString(onCallMailBody(t, mail.Header(part.Header), part))
		}
		return content.String()
	}
	switch strings.ToLower(header.Get("Content-Transfer-Encoding")) {
	case "base64":
		body = base64.NewDecoder(base64.StdEncoding, body)
	case "quoted-printable":
		body = quotedprintable.NewReader(body)
	}
	content, err := io.ReadAll(body)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func TestOnCallSMTPDeliveryWithPostgres(t *testing.T) {
	for _, scenario := range []struct {
		name       string
		reject     bool
		logFailure bool
		recovery   bool
	}{{name: "accepted"}, {name: "recovery", recovery: true}, {name: "rejected", reject: true}, {name: "log-failed", logFailure: true}} {
		t.Run(scenario.name, func(t *testing.T) {
			reject := scenario.reject
			db := onCallScenarioDB(t)
			address, received := onCallSMTPReceiver(t, reject)
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				t.Fatal(err)
			}
			// Replace the complete cache, since configuration is process-global even
			// though this test's database schema is isolated.
			configValueCacheMu.Lock()
			oldCache := configValueCache
			configValueCache = make(map[string]string)
			configValueCacheMu.Unlock()
			t.Cleanup(func() { configValueCacheMu.Lock(); configValueCache = oldCache; configValueCacheMu.Unlock() })
			configs := map[string]string{"email_from": "sender@example.test", "email_nickname": "Loopback test", "email_secret": "loopback-only-secret", "email_host": host, "email_port": port, "email_isSSl": "false", "encryption_key": "oncall-test-encryption-key"}
			for key, value := range configs {
				if err := db.Create(&Config{ConfigKey: key, ConfigValue: value}).Error; err != nil {
					t.Fatal(err)
				}
			}
			cwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			templates, err := filepath.Abs(filepath.Join(cwd, "../../assets/templates"))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(filepath.Join(templates, "mail_problem.tpl")); err != nil {
				t.Fatal(err)
			}
			t.Chdir(filepath.Dir(filepath.Dir(templates)))
			user := User{ID: 101, Username: "smtp-oncall", Email: "oncall@example.test"}
			if err := db.Create(&user).Error; err != nil {
				t.Fatal(err)
			}
			now := time.Now()
			schedule := OnCallSchedule{Name: "SMTP integration", Timezone: "Asia/Shanghai", Rotation: "fixed", Start: now.Add(-24 * time.Hour).Format(time.RFC3339), Members: []int{user.ID}, Weekdays: []int{0, 1, 2, 3, 4, 5, 6}, WindowStart: "00:00", WindowEnd: "00:00", Enabled: true}
			if err := SaveOnCallSchedule(&schedule, "smtp-test"); err != nil {
				t.Fatal(err)
			}
			rule := Rule{Name: "SMTP CPU routing", ZIDs: "41", MType: "1", Status: "0", Channel: "mail", ScheduleIDs: []int{schedule.ID}, Stime: "00:00", Etime: "23:59", Sweek: "0,1,2,3,4,5,6", Conditions: `[{"r_type":"key","r_func":"==","r_value":"system.cpu.util"}]`}
			if _, err := AddRule(&rule); err != nil {
				t.Fatal(err)
			}
			alarm := Alarm{ZID: 41, EventID: 81001, Host: "smtp-test-host", Hostname: "SMTP test server", HostsIP: "192.0.2.10", Hkey: "system.cpu.util", ItemName: "CPU utilization", ItemValue: "95.5", Message: "CPU usage threshold breached", Detail: "CPU utilization:95.5", Level: "4", Status: "1", OccurTime: now}
			if scenario.recovery {
				alarm.Status = "0"
				alarm.Message = "CPU usage recovered"
			}
			if _, err := AddAlarm(&alarm); err != nil {
				t.Fatal(err)
			}
			GenAlert(&alarm)
			events := PopAllMail()
			if len(events) != 1 || events[0].ToUsers != strconv.Itoa(user.ID) {
				t.Fatalf("unexpected mail queue: %+v", events)
			}
			if scenario.logFailure {
				if err := db.Migrator().DropTable(&EventLog{}); err != nil {
					t.Fatal(err)
				}
			}
			if reject || scenario.logFailure {
				err := SendEmailAlert(events[0], user)
				if err == nil {
					t.Error("delivery or log failure must be returned to the caller")
				}
				if reject && (err == nil || !strings.Contains(err.Error(), "550")) {
					t.Errorf("missing SMTP rejection: %v", err)
				}
				if scenario.logFailure && (err == nil || !strings.Contains(err.Error(), "event_log")) {
					t.Errorf("missing delivery log database error: %v", err)
				}
			} else {
				oldWorker := MailWorkerChan
				MailWorkerChan = make(chan int, 1)
				t.Cleanup(func() { MailWorkerChan = oldWorker })
				MailWorkerChan <- 1
				SendMail(events[0])
			}
			if events[0].Status != alarm.Status || events[0].Level != alarm.Level {
				t.Errorf("sending mutated queued event: status=%q level=%q", events[0].Status, events[0].Level)
			}
			var delivered onCallSMTPMessage
			select {
			case delivered = <-received:
			case <-time.After(5 * time.Second):
				t.Fatal("no SMTP DATA received")
			}
			if delivered.from != "<sender@example.test>" || delivered.to != "<oncall@example.test>" {
				t.Fatalf("wrong SMTP envelope: from=%s to=%s", delivered.from, delivered.to)
			}
			message, err := mail.ReadMessage(strings.NewReader(delivered.raw))
			if err != nil {
				t.Fatal(err)
			}
			subject, err := new(mime.WordDecoder).DecodeHeader(message.Header.Get("Subject"))
			if err != nil || !strings.Contains(subject, alarm.Message) {
				t.Fatalf("wrong subject %q: %v", subject, err)
			}
			wantStatus := "故障"
			if scenario.recovery {
				wantStatus = "恢复"
			}
			if !strings.HasPrefix(subject, "["+wantStatus+"]") {
				t.Errorf("wrong display status in subject: %q", subject)
			}
			body := onCallMailBody(t, message.Header, message.Body)
			if strings.Contains(body, "Problem resolved:") != scenario.recovery {
				t.Errorf("wrong problem/recovery template for status %q", alarm.Status)
			}
			if !strings.Contains(body, "High") {
				t.Error("mail body lost severity display label")
			}

			for _, value := range []string{alarm.Message, alarm.ItemValue, alarm.Hostname, alarm.HostsIP} {
				if !strings.Contains(body, value) {
					t.Errorf("mail body does not contain %q", value)
				}
			}
			if scenario.logFailure {
				return
			}
			var logs []EventLog
			if err := db.Where("alarm_id = ?", alarm.ID).Find(&logs).Error; err != nil {
				t.Fatal(err)
			}
			expected := strconv.Itoa(EventSuccess)
			if reject {
				expected = strconv.Itoa(EventFailed)
			}
			if len(logs) != 1 || logs[0].Status != expected || logs[0].User != user.Username || logs[0].Account != user.Email || logs[0].EventID != alarm.EventID {
				t.Fatalf("wrong delivery log: %+v", logs)
			}
			if reject && !strings.Contains(logs[0].NotifyError, "550") {
				t.Fatalf("SMTP rejection missing in log: %+v", logs[0])
			}
		})
	}
}
