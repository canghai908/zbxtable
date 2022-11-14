package models

import "time"

//TableName Topology
func (t *TopologyData) TableName() string {
	return TableName("topology_data")
}

type TopologyData struct {
	ID        int       `orm:"column(id);auto" json:"id"`
	PID       string    `orm:"column(pid);size(60)" json:"pid"`
	VType     string    `orm:"column(v_type);size(60)" json:"v_type"`
	Type      string    `orm:"column(type);size(60)" json:"type"`
	TID       string    `orm:"column(tid);size(60)" json:"tid"`
	CreatedAt time.Time `orm:"column(created_at);auto_now_add;type(datetime)" json:"created_at"`
	UpdatedAt time.Time `orm:"column(updated_at);auto_now;type(datetime)" json:"updated_at"`
}

type AllEdge []AEdge

type AEdge struct {
	Attrs struct {
		Line struct {
			FlowID       string `json:"FlowID"`
			FlowName     string `json:"FlowName"`
			FlowType     string `josn:"FlowType""`
			HostID       string `json:"HostID"`
			HostName     string `json:"HostName"`
			HostType     string `json:"HostType"`
			HostValue    string `json:"HostValue"`
			TriggerDesc  string `json:"TriggerDesc"`
			TriggerID    string `json:"TriggerID"`
			DataSource   string `json:"dataSource"`
			InTape       string `json:"inTape"`
			OutTape      string `json:"outTape"`
			Stroke       string `json:"stroke"`
			StrokeWidth  int    `json:"strokeWidth"`
			SourceMarker struct {
				Name   string `json:"name"`
				Size   int64  `json:"size"`
				Stroke string `json:"stroke"`
			} `json:"sourceMarker"`
			StrokeDasharray int64 `json:"strokeDasharray"`
			Style           struct {
				Animation string `json:"animation"`
			} `json:"style"`
			TargetMarker struct {
				Name string `json:"name"`
				Size int64  `json:"size"`
			} `json:"targetMarker"`
		} `json:"line"`
	} `json:"attrs"`
	Connector string `json:"connector"`
	ID        string `json:"id"`
	Labels    []struct {
		Attrs struct {
			Label struct {
				Text string `json:"text"`
			} `json:"label"`
		} `json:"attrs"`
		Position struct {
			Distance string  `json:"distance"`
			Offset   float64 `json:"offset"`
			Angle    int     `json:"angle"`
			Options  struct {
				KeepGradient     bool `json:"keepGradient"`
				EnsureLegibility bool `json:"ensureLegibility"`
			} `json:"options"`
		} `json:"position"`
	} `json:"labels"`
	Router struct {
		Name string `json:"name"`
	} `json:"router"`
	Shape  string `json:"shape"`
	Source struct {
		Cell string `json:"cell"`
		Port string `json:"port"`
	} `json:"source"`
	Target struct {
		Cell string `json:"cell"`
		Port string `json:"port"`
	} `json:"target"`
	ZIndex int64 `json:"zIndex"`
}

type AllNodes []PNodes

type PNodes struct {
	Attrs struct {
		Image struct {
			Xlink_href string `json:"xlink:href"`
		} `json:"image"`
		Label struct {
			HostID     string `json:"HostID"`
			HostType   string `json:"HostType"`
			HostValue  string `json:"HostValue"`
			HostCPU    string `json:"HostCPU"`
			HostMem    string `json:"HostMem"`
			HostAlarm  string `json:"HostAlarm"`
			HostClock  string `json:"HostClock"`
			HostStatus string `json:"HostStatus"`
			HostError  string `json:"HostError"`
			Text       string `json:"text"`
		} `json:"label"`
		Text struct {
			Text string `json:"text"`
		} `json:"text"`
	} `json:"attrs"`
	ID    string `json:"id"`
	Ports struct {
		Groups struct {
			Bottom struct {
				Attrs struct {
					Circle struct {
						Fill        string `json:"fill"`
						Magnet      bool   `json:"magnet"`
						R           int64  `json:"r"`
						Stroke      string `json:"stroke"`
						StrokeWidth int64  `json:"strokeWidth"`
					} `json:"circle"`
				} `json:"attrs"`
				Position string `json:"position"`
			} `json:"bottom"`
			Left struct {
				Attrs struct {
					Circle struct {
						Fill        string `json:"fill"`
						Magnet      bool   `json:"magnet"`
						R           int64  `json:"r"`
						Stroke      string `json:"stroke"`
						StrokeWidth int64  `json:"strokeWidth"`
					} `json:"circle"`
				} `json:"attrs"`
				Position string `json:"position"`
			} `json:"left"`
			Right struct {
				Attrs struct {
					Circle struct {
						Fill        string `json:"fill"`
						Magnet      bool   `json:"magnet"`
						R           int64  `json:"r"`
						Stroke      string `json:"stroke"`
						StrokeWidth int64  `json:"strokeWidth"`
					} `json:"circle"`
				} `json:"attrs"`
				Position string `json:"position"`
			} `json:"right"`
			Top struct {
				Attrs struct {
					Circle struct {
						Fill        string `json:"fill"`
						Magnet      bool   `json:"magnet"`
						R           int64  `json:"r"`
						Stroke      string `json:"stroke"`
						StrokeWidth int64  `json:"strokeWidth"`
					} `json:"circle"`
				} `json:"attrs"`
				Position string `json:"position"`
			} `json:"top"`
		} `json:"groups"`
		Items []struct {
			Group string `json:"group"`
			ID    string `json:"id"`
		} `json:"items"`
		//Style struct {
		//	Visibility bool `json:"visibility"`
		//} `json:"style"`
	} `json:"ports"`
	Position struct {
		X int64 `json:"x"`
		Y int64 `json:"y"`
	} `json:"position"`
	Shape string `json:"shape"`
	Size  struct {
		Height int64 `json:"height"`
		Width  int64 `json:"width"`
	} `json:"size"`
	ZIndex int64 `json:"zIndex"`
}
