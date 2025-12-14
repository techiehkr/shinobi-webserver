; installer-windows.iss
[Setup]
AppName=Shinobi Web Server
AppVersion=1.0.0
DefaultDirName={pf}\ShinobiWebServer
DefaultGroupName=Shinobi Web Server
UninstallDisplayIcon={app}\shinobi-webserver.exe
Compression=lzma2
SolidCompression=yes
OutputDir=.\dist
OutputBaseFilename=ShinobiWebServer-Setup

[Files]
Source: "build\windows\shinobi-webserver.exe"; DestDir: "{app}"
Source: "LICENSE"; DestDir: "{app}"
Source: "README.md"; DestDir: "{app}"
Source: "assets\*"; DestDir: "{app}\assets"; Flags: recursesubdirs

[Icons]
Name: "{group}\Shinobi Web Server"; Filename: "{app}\shinobi-webserver.exe"
Name: "{group}\Uninstall Shinobi"; Filename: "{uninstallexe}"
Name: "{commondesktop}\Shinobi Web Server"; Filename: "{app}\shinobi-webserver.exe"

[Run]
Filename: "{app}\shinobi-webserver.exe"; Description: "Launch Shinobi Web Server"; Flags: postinstall nowait skipifsilent