%bcond_without srpm


Name:       gonzofilter
Version:    0.5.0
Release:    1%{?dist}
Summary:    Bayes classifying spam mail filter
URL:        https://github.com/gsauthof/gonzofilter
License:    GPLv3+
Source:     https://example.org/gonzofilter.tar


BuildRequires: golang-etcd-bbolt-devel
BuildRequires: golang-x-sys-devel
BuildRequires: golang-x-text-devel

%description
Bayes classifying spam mail filter

%prep
%if %{with srpm}
%autosetup -n gonzofilter
%endif

%build
GOPATH=$HOME/go:/usr/share/gocode go build

%install
mkdir -p %{buildroot}/usr/bin
cp gonzofilter %{buildroot}/usr/bin

%check
GOPATH=$HOME/go:/usr/share/gocode go test -v

%files
/usr/bin/gonzofilter
%doc README.md


%changelog
* Thu Sep 17 2020 Georg Sauthoff <mail@gms.tf> - 0.5.0-1
- initial packaging

