# Structured Access Zone (saz)

<img align="right" src="saz.png" alt="saz" width="184">

[![License](https://img.shields.io/github/license/worldline-go/saz?color=red&style=flat-square)](https://raw.githubusercontent.com/worldline-go/saz/main/LICENSE)
[![Coverage](https://img.shields.io/sonar/coverage/worldline-go_saz?logo=sonarcloud&server=https%3A%2F%2Fsonarcloud.io&style=flat-square)](https://sonarcloud.io/summary/overall?id=worldline-go_saz)
[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/worldline-go/saz/test.yml?branch=main&logo=github&style=flat-square&label=ci)](https://github.com/worldline-go/saz/actions)

Helps to transfer data from one database to another database and play queries🎶.

__-__ Postgres(`pgx`), Oracle(`godror`), SQLite3(`sqlite3`), MSSQL(`sqlserver`), MySQL(`mysql`) supported  
__-__ UI for query writing  
__-__ Flexible data mapping  
__-__ Fast transfer with batch processing  
__-__ Plectrum of data  

<hr>

# Usage

Quick start with creating postgresql and run service.

```sh
# create postgresql
make env
# run service
make run
```

Open [http://localhost:8080](http://localhost:8080) to access the service and start writing queries.
