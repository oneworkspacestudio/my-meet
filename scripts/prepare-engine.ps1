param([string]$Cache='')
$ErrorActionPreference='Stop'
$root=Split-Path $PSScriptRoot -Parent
if(-not $Cache){$Cache=if($env:MYMEET_ENGINE_CACHE){$env:MYMEET_ENGINE_CACHE}else{Join-Path $root 'dist/engine-cache'}}
$pin=Get-Content -Raw -LiteralPath (Join-Path $root 'engine.lock.json')|ConvertFrom-Json
New-Item -ItemType Directory -Force -Path $Cache|Out-Null
$archive=Join-Path $Cache $pin.asset
if(-not(Test-Path -LiteralPath $archive)){
 Invoke-WebRequest -Uri ('https://github.com/'+$pin.repository+'/releases/download/'+$pin.tag+'/'+$pin.asset) -OutFile $archive
}
if((Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash.ToLowerInvariant() -cne $pin.sha256){throw 'LiveKit archive checksum mismatch'}
Add-Type -AssemblyName System.IO.Compression.FileSystem
$zip=[IO.Compression.ZipFile]::OpenRead($archive)
try{
 $entry=$zip.GetEntry('livekit-server.exe')
 if(-not $entry){throw 'LiveKit executable missing'}
 $mem=[IO.MemoryStream]::new();$input=$entry.Open()
 try{$input.CopyTo($mem)}finally{$input.Dispose()}
 $bytes=$mem.ToArray()
 $actual=[Convert]::ToHexString([Security.Cryptography.SHA256]::HashData($bytes)).ToLowerInvariant()
 if($actual -cne $pin.exeSHA256){throw 'LiveKit executable checksum mismatch'}
 $dest=Join-Path $root 'internal/engine/payload'
 New-Item -ItemType Directory -Force -Path $dest|Out-Null
 $out=[IO.File]::Create((Join-Path $dest 'livekit.exe.gz'))
 try{$gz=[IO.Compression.GZipStream]::new($out,[IO.Compression.CompressionLevel]::Optimal,$true);try{$gz.Write($bytes)}finally{$gz.Dispose()}}finally{$out.Dispose();$mem.Dispose()}
}finally{$zip.Dispose()}
Write-Output 'Prepared the verified LiveKit Windows engine.'
