Name: hydra-worker-map-sort
Version: 1
Release: 0
Summary: hydra-worker-map-sort
Source0: hydra-worker-map-sort-1.0.tar.gz
License: MIT
Group: custom
URL: https://github.com/innotech/hydra-worker-map-sort
BuildArch: x86_64
BuildRoot: %{_tmppath}/%{name}-buildroot
Requires: libzmq3
%description
Map instances by attribute and sort them.
%prep
%setup -q
%build
%install
install -m 0755 -d $RPM_BUILD_ROOT/usr/local/hydra
install -m 0755 hydra-worker-map-sort $RPM_BUILD_ROOT/usr/local/hydra/hydra-worker-map-sort

install -m 0755 -d $RPM_BUILD_ROOT/etc/init.d
install -m 0755 hydra-worker-map-sort-init.d.sh $RPM_BUILD_ROOT/etc/init.d/hydra-worker-map-sort

install -m 0755 -d $RPM_BUILD_ROOT/etc/hydra
install -m 0644 hydra.conf $RPM_BUILD_ROOT/etc/hydra/hydra-worker-map-sort.conf
%clean
rm -rf $RPM_BUILD_ROOT
%post
echo   You should edit config file /etc/hydra/hydra-worker-map-sort.conf
echo   When finished, you may want to run \"update-rc.d hydra-worker-map-sort defaults\"
%files
/usr/local/hydra/hydra-worker-map-sort
/etc/init.d/hydra-worker-map-sort
