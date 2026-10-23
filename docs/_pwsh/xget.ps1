function Invoke-XGet {
    <#
    .EXTERNALHELP xget-help.xml
    #>
    $application = Get-Command xget -CommandType Application -ErrorAction Stop | Select-Object -First 1
    & $application @args
    $global:LASTEXITCODE = $LASTEXITCODE
}

function Invoke-XGetInstall {
    <#
    .EXTERNALHELP xget-help.xml
    #>
    Invoke-XGet install @args
}

function Invoke-XGetUninstall {
    <#
    .EXTERNALHELP xget-help.xml
    #>
    Invoke-XGet uninstall @args
}

function Invoke-XGetUpgrade {
    <#
    .EXTERNALHELP xget-help.xml
    #>
    Invoke-XGet upgrade @args
}

function Invoke-XGetSelfUpdate {
    <#
    .EXTERNALHELP xget-help.xml
    #>
    Invoke-XGet self-update @args
}

function Invoke-XGetList {
    <#
    .EXTERNALHELP xget-help.xml
    #>
    Invoke-XGet list @args
}

function Invoke-XGetRate {
    <#
    .EXTERNALHELP xget-help.xml
    #>
    Invoke-XGet rate @args
}

function Invoke-XGetVersion {
    <#
    .EXTERNALHELP xget-help.xml
    #>
    Invoke-XGet version @args
}

function Invoke-XGetConfig {
    <#
    .EXTERNALHELP xget-help.xml
    #>
    Invoke-XGet config @args
}

function Invoke-XGetConfigGet {
    <#
    .EXTERNALHELP xget-help.xml
    #>
    Invoke-XGet config get @args
}

function Invoke-XGetConfigSet {
    <#
    .EXTERNALHELP xget-help.xml
    #>
    Invoke-XGet config set @args
}

function Invoke-XGetConfigClear {
    <#
    .EXTERNALHELP xget-help.xml
    #>
    Invoke-XGet config clear @args
}

function Invoke-XGetConfigPop {
    <#
    .EXTERNALHELP xget-help.xml
    #>
    Invoke-XGet config pop @args
}

function Invoke-XGetConfigList {
    <#
    .EXTERNALHELP xget-help.xml
    #>
    Invoke-XGet config list @args
}

function Invoke-XGetConfigPath {
    <#
    .EXTERNALHELP xget-help.xml
    #>
    Invoke-XGet config path @args
}

function Invoke-XGetConfigEdit {
    <#
    .EXTERNALHELP xget-help.xml
    #>
    Invoke-XGet config edit @args
}

function Invoke-XGetCompletion {
    <#
    .EXTERNALHELP xget-help.xml
    #>
    Invoke-XGet completion @args
}

function Invoke-XGetHelp {
    <#
    .EXTERNALHELP xget-help.xml
    #>
    Invoke-XGet help @args
}

function Invoke-XGetCompletionBash {
    <#
    .EXTERNALHELP xget-help.xml
    #>
    Invoke-XGet completion bash @args
}

function Invoke-XGetCompletionZsh {
    <#
    .EXTERNALHELP xget-help.xml
    #>
    Invoke-XGet completion zsh @args
}

function Invoke-XGetCompletionFish {
    <#
    .EXTERNALHELP xget-help.xml
    #>
    Invoke-XGet completion fish @args
}

function Invoke-XGetCompletionPowershell {
    <#
    .EXTERNALHELP xget-help.xml
    #>
    Invoke-XGet completion powershell @args
}

Set-Alias -Name xget -Value Invoke-XGet
Set-Alias -Name xget-install -Value Invoke-XGetInstall
Set-Alias -Name xget-uninstall -Value Invoke-XGetUninstall
Set-Alias -Name xget-remove -Value Invoke-XGetUninstall
Set-Alias -Name xget-upgrade -Value Invoke-XGetUpgrade
Set-Alias -Name xget-update -Value Invoke-XGetUpgrade
Set-Alias -Name xget-self-update -Value Invoke-XGetSelfUpdate
Set-Alias -Name xget-list -Value Invoke-XGetList
Set-Alias -Name xget-rate -Value Invoke-XGetRate
Set-Alias -Name xget-version -Value Invoke-XGetVersion
Set-Alias -Name xget-config -Value Invoke-XGetConfig
Set-Alias -Name xget-config-get -Value Invoke-XGetConfigGet
Set-Alias -Name xget-config-set -Value Invoke-XGetConfigSet
Set-Alias -Name xget-config-clear -Value Invoke-XGetConfigClear
Set-Alias -Name xget-config-pop -Value Invoke-XGetConfigPop
Set-Alias -Name xget-config-list -Value Invoke-XGetConfigList
Set-Alias -Name xget-config-path -Value Invoke-XGetConfigPath
Set-Alias -Name xget-config-edit -Value Invoke-XGetConfigEdit
Set-Alias -Name xget-completion -Value Invoke-XGetCompletion
Set-Alias -Name xget-help -Value Invoke-XGetHelp
Set-Alias -Name xget-completion-bash -Value Invoke-XGetCompletionBash
Set-Alias -Name xget-completion-zsh -Value Invoke-XGetCompletionZsh
Set-Alias -Name xget-completion-fish -Value Invoke-XGetCompletionFish
Set-Alias -Name xget-completion-powershell -Value Invoke-XGetCompletionPowershell