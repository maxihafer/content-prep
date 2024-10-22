# content-prep - Intune packages in pure go

## Description

This project aims to allow packaging Win32 apps for use with Intune.

## Usage

### Binary

```shell
content-prep new --path "path/to/source" --setupFile "path/to/source/setupFile" --output "path/to/output"
```

### Docker
```shell
docker run ghcr.io/maxihafer/content-prep:latest \
  -v /path/to/source:/src \
  -v /path/to/output:/out \
  content-prep new --path /src --setupFile /src/setupFile --output /out
```


## Motivation
 Microsoft provides its closed-source [Content-Prep-Tool](https://github.com/microsoft/Microsoft-Win32-Content-Prep-Tool) for packaging Applications for intune. 
 Being just a checked in `.exe` file, it is neither possible to verify the code nor its behavior.

There exist great open-source projects that implement the same functionality. Even though some of them work cross-platform, they still rely on `.NET` and `PowerShell` which is not ideal for the use case of easily publishing intune packages from CI/CD.
 
The great work of [@svrooij](https://github.com/svrooij) really helped alot in understanding the inner workings of Intune packages:

- [Content-Prep](https://github.com/svrooij/ContentPrep) - `.NET/PowerShell` implementation of the Content-Prep functionality.
- [WinTuner](https://github.com/svrooij/wingetintune) - Very sophisticated `PowerShell` Module for creating, deploying and managing Intune packages.
- [The Intune series in his blog](https://svrooij.io/2023/08/30/apps-intune/) - Great insights into the inner workings of `.intunewin` packages. Very well written and concise.

## Intune package structure

A `.intunewin` package is constructed as follows:
```shell
foo.intunewin (zip)
└── IntuneWinPackage
    ├── Contents
    │   └── IntunePackage.intunewin
    └── Metadata
        └── Detection.xml
```

### `IntunePackage.intunewin`
This is a encrypted `.zip` file containing the actual content of the package. More info can be found in the [Encryption Section](#encryption-process).

### `foo.intunewin/Metadata/Detection.xml`
The `Detection.xml` file is read by Intune to extract all kinds of information about the package. It is a `XML` file with the following structure:
```xml
<ApplicationInfo ToolVersion="1.8.4.0">
    <FileName>IntunePackage.intunewin</FileName>                                    // This is the static name of the inner (encrypted) package
    <Name>foo</Name>                                                                // Display name of the application
    <UnencryptedContentSize>216404340</UnencryptedContentSize>                      // Size of the unencrypted content
    <SetupFile>foo.exe</SetupFile>                                                  // Name of the setup file
    <EncryptionInfo>
        <EncryptionKey>QX7BMsPzNDBSWv836oYT+gpxKlDYyPDm6C55XZKm9Z8=</EncryptionKey> // Base64 encoded encryption key (32 byte)
        <MacKey>Ysr5oRUvkgORiurV3zCf14jsj/baZOE/Xc6EEm/2wYA=</MacKey>               // Base64 encoded HMAC key (32 byte)
        <InitializationVector>/Nh7KHI5lYFyCTbGqBASPg==</InitializationVector>       // Base64 encoded IV (16 byte)
        <Mac>PmGnbIzb6/N4pc3zZJF70+PYEAkXezR9Q6PaC4CzBdY=</Mac>                     // Base64 encoded HMAC
        <ProfileIdentifier>ProfileVersion1</ProfileIdentifier>                      // Static encoding Profile used by Intune to verify package integrity
        <FileDigest>47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=</FileDigest>       // Base64 encoded SHA256 hash of the encrypted content
        <FileDigestAlgorithm>SHA256</FileDigestAlgorithm>                           // Hashing algorithm used
    </EncryptionInfo>
</ApplicationInfo>
```

## Encryption Process

1. The source folder (`src`) is zipped without any compression to a temporary file.
2. The output file (`out`) is created and its write position is advanced by the length of the HMAC.
3. The IV is written to the output file (at position `0 + len(HMAC)`) as well as the HMAC itself.
4. The zipped `src` is then XORKeyStreamed into the output file starting at position `0 + len(HMAC) + len(IV)` as well as into the HMAC.
5. The HMAC is then calculated and written to the output file at position `0`.
6. The metadata is written to the `Detection.xml` file.<br/>
**_NOTE:_** the digest changes from execution to execution,. This is due to the fact that the zip archive contains the file creation time in its header.

The resulting structure of the encrypted file is as follows:

| 32 Bytes                       | 16 Bytes | AES-CTR XORKeyStreamed Content |
|--------------------------------|----------|--------------------------------|
| HMAC of IV + Encrypted Content | IV       | Encrypted Content              |
