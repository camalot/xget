Name:           xget
Version:        __VERSION__
Release:        1%{?dist}
Summary:        Download pre-built binaries from GitHub and GitLab releases

License:         MIT
URL:             https://github.com/camalot/xget
Source0:         %{name}-%{version}.tar.gz
BuildRequires:   golang >= 1.27
BuildRequires:   pandoc

%description
xget downloads and extracts pre-built binaries from GitHub and GitLab releases.
It selects a suitable release asset for the current platform, verifies available
checksums, and installs the requested executable locally.

%prep
%autosetup

%build
CGO_ENABLED=0 go build -buildvcs=false \
	-ldflags="-s -w -X github.com/camalot/xget/internal/cli.version=%{version}" \
	-o xget ./cmd/xget
pandoc docs/_man/xget.md -s -t man -o xget.1

%install
install -Dpm 0755 xget %{buildroot}%{_bindir}/xget
install -Dpm 0644 xget.1 %{buildroot}%{_mandir}/man1/xget.1
install -Dpm 0644 LICENSE %{buildroot}%{_licensedir}/%{name}/LICENSE
install -Dpm 0644 README.md %{buildroot}%{_docdir}/%{name}/README.md

%check
%{buildroot}%{_bindir}/xget --version

%files
%license LICENSE
%doc README.md
%{_bindir}/xget
%{_mandir}/man1/xget.1*

%changelog
* Thu Sep 24 2026 Ryan Conrad <camalot@gmail.com> - __VERSION__-1
- Package xget
