About
======

Push vdbench realtime metrics to Prometheus Pushgateway.

Usage
-----

- Linux

  ::

    cp vdbench_pusher path/to/vdbench
    ./vdbench <options> | ./vdbench_pusher --gateway http://localhost:9091 --job test1

- Windows

  **Notes**: Classic CMD console should be used. The PowerShell console buffers all vdbench output as objects and won't pass result line by line which lead to weird behavior.

  ::

    cp vdbench_pusher path\to\vdbench
    .\vdbench.bat <options> | .\vdbench_pusher.exe --gateway http://localhost:9091 --job test1
