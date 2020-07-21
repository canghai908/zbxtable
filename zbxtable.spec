Name:		zbxtable
Version:	1.0.0
Release: 	1%{?alphatag:.%{alphatag}}%{?dist}
Summary:	A tools export table on Zabbix
Group:		Applications/Internet
License:	Apache-2.0
URL:		https://blog.cactifans.com/
Source0:	zbxtable-%{version}%{?alphatag:%{alphatag}}.tar.gz
Requires(pre):		/usr/sbin/useradd

Buildroot:	%{_tmppath}/zbxtable-%{version}-%{release}-root-%(%{__id_u} -n)

%description
A tools send zabbix alerts to ZbxTable

%global debug_package %{nil}

%prep
%setup0 -q -n zbxtable-%{version}%{?alphatag:%{alphatag}}

%build

%install

rm -rf $RPM_BUILD_ROOT

# install necessary directories
mkdir -p $RPM_BUILD_ROOT%{_prefix}/local/zbxtable
mkdir -p $RPM_BUILD_ROOT%{_prefix}/local/zbxtable/conf
mkdir -p $RPM_BUILD_ROOT%{_prefix}/local/zbxtable/keys

# install  binaries and conf
install -m 0755 -p zbxtable $RPM_BUILD_ROOT%{_prefix}/local/zbxtable/
install -m 0755 -p nginx.conf $RPM_BUILD_ROOT%{_prefix}/local/zbxtable/
install -m 0755 -p msty.ttf $RPM_BUILD_ROOT%{_prefix}/local/zbxtable/
install -m 0755 -p conf/app.conf $RPM_BUILD_ROOT%{_prefix}/local/zbxtable/conf/
install -m 0755 -p keys/rsakey.pem $RPM_BUILD_ROOT%{_prefix}/local/zbxtable/keys/
install -m 0755 -p keys/rsakey.pem.pub $RPM_BUILD_ROOT%{_prefix}/local/zbxtable/keys/

#systemd
install -Dm 0644 -p zbxtable.service $RPM_BUILD_ROOT%{_unitdir}/zbxtable.service

exit 0

%clean
rm -rf $RPM_BUILD_ROOT

%pre
getent group zbxtable > /dev/null || groupadd -r zbxtable
getent passwd zbxtable > /dev/null || \
	useradd -r -g zbxtable -d %{_localstatedir}/lib/zbxtable -s /sbin/nologin \
	-c "ZbxTable System" zbxtable
:


%define __debug_install_post   \
   %{_rpmconfigdir}/find-debuginfo.sh %{?_find_debuginfo_opts} "%{_builddir}/%{?buildsubdir}"\
%{nil}

%files
%defattr(755,zbxtable,zbxtable,755)
%attr(0755,zbxtable,zbxtable)
%dir %{_prefix}/local/zbxtable
%dir %{_prefix}/local/zbxtable/conf
%dir %{_prefix}/local/zbxtable/keys
%{_prefix}/local/zbxtable/zbxtable
%{_prefix}/local/zbxtable/nginx.conf
%{_prefix}/local/zbxtable/msty.ttf
%{_prefix}/local/zbxtable/conf/app.conf
%{_prefix}/local/zbxtable/keys/rsakey.pem
%{_prefix}/local/zbxtable/keys/rsakey.pem.pub
%{_unitdir}/zbxtable.service