Name:		zbxtable
Version:	%{version}
Release: 	1%{?alphatag:.%{alphatag}}%{?dist}
Summary:	A tools export table on Zabbix
Group:		Applications/Internet
License:	Apache-2.0
URL:		https://zbxtable.cactifans.com
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

# install  binaries and conf
install -m 0755 -p zbxtable $RPM_BUILD_ROOT%{_prefix}/local/zbxtable/

#install startup scripts
%if 0%{?rhel} >= 7
install -Dm 0755 -p zbxtable.service $RPM_BUILD_ROOT%{_unitdir}/zbxtable.service
%else
install -Dm 0755 -p zbxtable.init $RPM_BUILD_ROOT%{_sysconfdir}/init.d/zbxtable
%endif
exit 0

%clean
rm -rf $RPM_BUILD_ROOT

%pre
getent group zbxtable > /dev/null || groupadd -r zbxtable
getent passwd zbxtable > /dev/null || \
	useradd -r -g zbxtable -d %{_localstatedir}/lib/zbxtable -s /sbin/nologin \
	-c "ZbxTable System" zbxtable
:

%post
%if 0%{?rhel} >= 7
%systemd_post zbxtable.service
%else
/sbin/chkconfig --add zbxtable
%endif
:


%define __debug_install_post   \
   %{_rpmconfigdir}/find-debuginfo.sh %{?_find_debuginfo_opts} "%{_builddir}/%{?buildsubdir}"\
%{nil}

%files
%defattr(755,zbxtable,zbxtable,755)
%attr(0755,zbxtable,zbxtable)
%dir %{_prefix}/local/zbxtable
%{_prefix}/local/zbxtable/zbxtable

%if 0%{?rhel} >= 7
%{_unitdir}/zbxtable.service
%else
%{_sysconfdir}/init.d/zbxtable
%endif