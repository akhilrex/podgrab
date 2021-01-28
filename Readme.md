<!--
*** Thanks for checking out this README Template. If you have a suggestion that would
*** make this better, please fork the repo and create a pull request or simply open
*** an issue with the tag "enhancement".
*** Thanks again! Now go create something AMAZING! :D
***
***
***
*** To avoid retyping too much info. Do a search and replace for the following:
*** akhilrex, podgrab, akhilrex, email
-->

<!-- PROJECT SHIELDS -->
<!--
*** I'm using markdown "reference style" links for readability.
*** Reference links are enclosed in brackets [ ] instead of parentheses ( ).
*** See the bottom of this document for the declaration of the reference variables
*** for contributors-url, forks-url, etc. This is an optional, concise syntax you may use.
*** https://www.markdownguide.org/basic-syntax/#reference-style-links
-->

[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![MIT License][license-shield]][license-url]
[![LinkedIn][linkedin-shield]][linkedin-url]

<!-- PROJECT LOGO -->
<br />
<p align="center">
  <!-- <a href="https://github.com/akhilrex/podgrab">
    <img src="images/logo.png" alt="Logo" width="80" height="80">
  </a> -->

  <h1 align="center">Podgrab</h1>

  <p align="center">
    A self-hosted podcast manager to download episodes as soon as they become live
    <br />
    <a href="https://github.com/akhilrex/podgrab"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <!-- <a href="https://github.com/akhilrex/podgrab">View Demo</a>
    · -->
    <a href="https://github.com/akhilrex/podgrab/issues">Report Bug</a>
    ·
    <a href="https://github.com/akhilrex/podgrab/issues">Request Feature</a>
        ·
    <a href="Screenshots.md">Screenshots</a>
  </p>
</p>

<!-- TABLE OF CONTENTS -->

## Table of Contents

- [About the Project](#about-the-project)
  - [Motivation](#motivation)
  - [Built With](#built-with)
- [Installation](#installation)
- [License](#license)
- [Roadmap](#roadmap)
- [Contact](#contact)

<!-- ABOUT THE PROJECT -->

## About The Project

Podgrab is a is a self-hosted podcast manager which automatically downloads latest podcast episodes. It is a light-weight application built using GO.

It works best if you already know which podcasts you want to monitor. However there is an experimental podcast search system powered by iTunes built into Podgrab

### Motivation

Podgrab started a tool that I initially built to solve a specific problem I had. During the COVID pandemic times I started going for a run. I do not prefer taking by phone along so I would add podcast episodes to my smart watch which could be connected with my bluetooth earphones. Most podcasting apps do not expose the mp3 files directly which is why I decided to build this quick tool for myself. Once it reached a stage where my requirements were fulfilled I decided to make it a little pretty and share it with everyone else.

![Product Name Screen Shot][product-screenshot]
[More Screenshots](Screenshots.md)

### Built With

- [Go](https://golang.org/)
- [Go-Gin](https://github.com/gin-gonic/gin)
- [GORM](https://github.com/go-gorm/gorm)
- [SQLite](https://www.sqlite.org/index.html)

## Installation

The easiest way to run Podgrab is to run it as a docker container.

### Using Docker

Simple setup without mounted volumes (for testing and evaluation)

```sh
  docker run -d -p 8080:8080 --name=podgrab akhilrex/podgrab
```

Binding local volumes to the container

```sh
   docker run -d -p 8080:8080 --name=podgrab -v "/host/path/to/assets:/assets" -v "/host/path/to/config:/config"  akhilrex/podgrab
```

### Using Docker-Compose

Modify the docker compose file provided [here](https://github.com/akhilrex/podgrab/blob/master/docker-compose.yml) to update the volume and port binding and run the following command

```yaml
version: "2.1"
services:
  podgrab:
    image: akhilrex/podgrab
    container_name: podgrab
    environment:
      - CHECK_FREQUENCY=240
    volumes:
      - /path/to/config:/config
      - /path/to/data:/assets
    ports:
      - 8080:8080
    restart: unless-stopped
```

```sh
   docker-compose up -d
```

### Environment Variables

| Name            | Description                                                             | Default |
| --------------- | ----------------------------------------------------------------------- | ------- |
| CHECK_FREQUENCY | How frequently to check for new episodes and missing files (in minutes) | 30      |
| PUID | Sets the UID of the container user | 1000 |
| PGID | Sets the GID of the container user | 1000 |

<!-- LICENSE -->

## License

Distributed under the MIT License. See `LICENSE` for more information.

## Roadmap

Following are the things that I plan to complete in the near future.

- Some more code refactoring.
- API standardisation so that it can be used to build apps on top of it.
- Better search and discovery

<!-- CONTACT -->

## Contact

Akhil Gupta - [@akhilrex](https://twitter.com/akhilrex)

Project Link: [https://github.com/akhilrex/podgrab](https://github.com/akhilrex/podgrab)

<a href="https://www.buymeacoffee.com/akhilrex" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/v2/default-yellow.png" alt="Buy Me A Coffee" style="height: 60px !important;width: 217px !important;" ></a>

<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->

[contributors-shield]: https://img.shields.io/github/contributors/akhilrex/podgrab.svg?style=flat-square
[contributors-url]: https://github.com/akhilrex/podgrab/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/akhilrex/podgrab.svg?style=flat-square
[forks-url]: https://github.com/akhilrex/podgrab/network/members
[stars-shield]: https://img.shields.io/github/stars/akhilrex/podgrab.svg?style=flat-square
[stars-url]: https://github.com/akhilrex/podgrab/stargazers
[issues-shield]: https://img.shields.io/github/issues/akhilrex/podgrab.svg?style=flat-square
[issues-url]: https://github.com/akhilrex/podgrab/issues
[license-shield]: https://img.shields.io/github/license/akhilrex/podgrab.svg?style=flat-square
[license-url]: https://github.com/akhilrex/podgrab/blob/master/LICENSE.txt
[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?style=flat-square&logo=linkedin&colorB=555
[linkedin-url]: https://linkedin.com/in/akhilrex
[product-screenshot]: images/screenshot.jpg
