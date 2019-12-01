# gcqlsh

Cassandra command line shell written in Golang

![](screenshots/gcqlsh_rec.gif?raw=true)

Motivation
----

Having a cassandra command line shell utility in one binary distributable.

Where it comes in hand?
----

 * Building docker images for cassandra from Alpine with no Python.
 * Running cql shell on all platforms.
 * Automating cassandra schema creation without need to install python dependencies.

Building
----

``` 
go build -o gocqlsh src/github.com/npenkov/gcqlsh/cmd/gcqlsh.go
```

Fatures
----
 * Running DDL script files from command line
 * Support for Cassandra 2.1+/ScyllaDB
 * CQL Support
 * Statement tracing
 * `desc` command with
   * `keyspaces` - simple list
   * `tables` - simple list
   * `table` - simple list of columns and types
 * Auto completition for commands:
   * `use` - keyspaces
   * `desc` - tables
   * `select` - tables
   * `update` - tables and columns
   * `delete` - tables
   * `insert` - tables

Still missing
----

 * Paging in interactive results
 * DDL Statements when describing Keyspaces and tables
 * Expanded rows
 * Code assistance for different keyspaces
 * Security when connecting to nodes
 * Node token awareness

Command line help
----

```
gcqlsh -h
Usage of gcqlsh:
  -f string
        Execute file containing cql statements instead of having interacive session
  -fail-on-error
        Stop execution if statement from file fails.
  -host string
        Cassandra host to connect to (default "127.0.0.1")
  -k string
        Default keyspace to connect to (default "system")
  -no-color
        Console without colors
  -port int
        Cassandra RPC port (default 9042)
  -print-confirmation
        Print 'ok' on successfuly executed cql statement from the file
  -print-cql
        Print Statements that are executed from a file
```

Planned features
----
 * `desc` for table
 * Column code assistance for 
   * `select`
   * `update`
   * `delete` 
   * `insert` 

Package dependencies
----

 * [Readline](https://github.com/chzyer/readline)
 * [Color](https://github.com/fatih/color)
 * [Gocql](https://github.com/gocql/gocql)

----

Written with [vim-go](https://github.com/fatih/vim-go)

License
-------

> Copyright (c) 2016-2019 Nick Penkov. All rights reserved.
> Use of this source code is governed by a MIT-style
> license that can be found in the LICENSE file.

