Name:           xget
Version:        __VERSION__
Release:        1%{?dist}
Summary:        Download pre-built binaries from GitHub and GitLab releases

License:         MIT
URL:             https://github.com/camalot/xget
Source0:         xget_%{version}_linux___ARCHITECTURE__.tar.gz

%description
xget downloads and extracts pre-built binaries from GitHub and GitLab releases.
It selects a suitable release asset for the current platform, verifies available
checksums, and installs the requested executable locally.

%prep
echo "__SHA256__  %{_sourcedir}/%{SOURCE0}" | sha256sum --check --strict
tar -xzf %{_sourcedir}/%{SOURCE0}

%build
# The release archive supplies the pre-built binary, manual page, and docs.

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
