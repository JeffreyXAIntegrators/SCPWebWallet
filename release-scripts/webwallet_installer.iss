; Script generated by the Inno Setup Script Wizard.
; SEE THE DOCUMENTATION FOR DETAILS ON CREATING INNO SETUP SCRIPT FILES!

#define MyAppName "ScPrime WebWallet"
#define MyAppVersion GetEnv('SCPVERSION')
#define MyAppPublisher "SCP Corp"
#define MyAppURL "https://www.scpri.me/"
#define MyAppExeName "scp-webwallet.exe"

[Setup]
; NOTE: The value of AppId uniquely identifies this application. Do not use the same AppId value in installers for other applications.
; (To generate a new GUID, click Tools | Generate GUID inside the IDE.)
AppId={{3909B164-9653-457B-BEEB-6C9CB0462539}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
;AppVerName={#MyAppName} {#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}
DefaultDirName={autopf}\{#MyAppName}
DisableProgramGroupPage=yes
LicenseFile=..\LICENSE
; Uncomment the following line to run in non administrative install mode (install for current user only.)
;PrivilegesRequired=lowest
PrivilegesRequiredOverridesAllowed=dialog
OutputDir=..\release-installer
OutputBaseFilename=webwallet_installer
Compression=lzma
SolidCompression=yes
WizardStyle=modern
ArchitecturesInstallIn64BitMode=x64
ArchitecturesAllowed=x64

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

[Files]
Source: "..\release\scp-webwallet-v{#MyAppVersion}-windows-amd64\{#MyAppExeName}"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\release\scp-webwallet-v{#MyAppVersion}-windows-amd64\scp-webwallet.exe.asc"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\release\scp-webwallet-server-v{#MyAppVersion}-windows-amd64\scp-webwallet-server.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\release\scp-webwallet-server-v{#MyAppVersion}-windows-amd64\scp-webwallet-server.exe.asc"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\release\scp-webwallet-v{#MyAppVersion}-windows-amd64\README.md"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\release\scp-webwallet-v{#MyAppVersion}-windows-amd64\LICENSE"; DestDir: "{app}"; Flags: ignoreversion
; NOTE: Don't use "Flags: ignoreversion" on any shared system files

[Icons]
Name: "{autoprograms}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon

[Run]
Filename: "{app}\{#MyAppExeName}"; Description: "{cm:LaunchProgram,{#StringChange(MyAppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent

