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

  **Notes**: Classic CMD console should be used. The PowerShell console buffers all vdbench output (as STDIN of the next command) and leads to unexpected processing behaviors (refer to `Powershell piping causes explosive memory usage<https://stackoverflow.com/questions/27440768/powershell-piping-causes-explosive-memory-usage>`_ for root cause).

  ::

    cp vdbench_pusher path\to\vdbench
    .\vdbench.bat <options> | .\vdbench_pusher.exe --gateway http://localhost:9091 --job test1
