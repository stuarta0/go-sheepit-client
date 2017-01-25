# (Go)SheepIt Render Farm Client

## Overview

**NOTE: This project is a work in progress and should not be used as a replacement for the official Java client at this time**

This is an unofficial clone of the [**SheepIt Render Farm Client**](https://github.com/laurent-clouet/sheepit-client) written in Go.

The purpose of this client is to provide a native executable for each platform that interacts with the distributed render farm [SheepIt](https://www.sheepit-renderfarm.com/). This removes the dependency on the JVM which reduces the required overhead and makes server deployment easier.

Feature-parity with the official Java client is a goal of this project. The same arguments that the Java client accepts can be provided to the Go client, however the ```-config <PATH>``` file structure has changed (see below).

## Differences to the Java client

* The Go client is CLI-only
* Configuration is now stored in TOML. The main difference to the INI file used by the Java client is that strings need to be quoted. To generate a configuration file automatically, run the Go client with the desired arguments with the additional arg ```-save-config <PATH>```.
* Additional arguments ```-project-dir <PATH>``` and ```-storage-dir <PATH>``` are provided for fine-grained control over project file location and Blender binary location respectively. Defaults are /tmp/ and ~/.sheepit/storage

## Compilation

    go get github.com/stuarta0/go-sheepit-client/...

## Usage

    gosheepit.exe -help

When you are doing development work, you can use a mirror of the main site specially made for demo/dev. The mirror is located at **http://sandbox.sheepit-renderfarm.com**, and you can use it by passing `-server http://sandbox.sheepit-renderfarm.com` to your invocation of the client.
