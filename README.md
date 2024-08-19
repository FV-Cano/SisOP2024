<h1 align="center">A.L.<span>GO</span></h1>

This project was developed for the subject Operative Systems (1st Semester 2024). It serves as a conceptualization of an operating system, breaking down its core functionalities into four distinct modules: Kernel, Memory, CPU, and IO Interfaces. These modules interact to provide a simplified representation of how an operating system functions.

These are some several key features implemented in this project:

- Long Term and Short Term Scheduling: Manages processes through FIFO, Round-Robin (RR), and Virtual Round-Robin (VRR) planning algorithms.
- Resource Management
- Fetch-Decode-Execute Cycle: Simulates the core CPU cycle, demonstrating how instructions are processed.
- Translation Lookaside Buffer (TLB): Optimizes memory access times through a caching mechanism.
- Memory Management (Pagination): Implements paging for efficient memory allocation and process management.
- Interfaces: Includes Generic, STDIN, STDOUT, and DIALFS interfaces to manage input, output, and file systems.
## Installation & Running

<ol>
  <li>Download the code by clicking the green button "Code" below the repository title and then selecting the option "Download ZIP".</li>
  <li>Extract the downloaded folder.</li>
  <li>If you are using a Windows OS, run the following files in order.
    <ol>
      <li>memoria.exe</li>
      <li>cpu.exe</li>
      <li>kernel.exe</li>
      <li>For IO devices, open a terminal inside entradasalida and run:</li>
    </ol>
  </li>

```bash
    ./entradasalida.exe config/<configname>.config
```
  Do this 'n' times being 'n' the amount of IO devices required for each test
  <li>Alternatively if you are using another OS, you have to build each module by running the command: "go build modulename.go" however, to do so you must have Golang installed in your system.</li>
  <li>It is highly recommended to have Postman installed for a better interaction with the program. To do so, go to https://www.postman.com/ and select the download option for your operative system. After installation, import the A.L.GO postman collection! You will be able to try all features with ease.</li>
</ol>

## Usage

You can play with the project by sending HTTP requests to the endpoints defined in the Postman collection, yet I recommend you to use the test cases in the section below.

## Running Tests

To run tests, first you must run some additional commands. Inside the algo-pruebas folder, open a terminal (you'll run the tests there) and run the following commands:

```bash
  export KERNEL_HOST=127.0.0.1
  export KERNEL_PORT=8001
```
After that, you can run the following commands for testing purposes. However, before running tests you must update config files inside the project. To see config file values for each test and expected values, visit the following document: https://docs.google.com/document/d/1XsBsJynoN5A9PTsTEaZsj0q3zsEtcnLgdAHOQ4f_4-g/edit

Inside scripts_kernel folder:
```bash
    ./<testname>.sh
```
Possible tests:

- PRUEBA_PLANI
- PRUEBA_IO
- PRUEBA_DEADLOCK
- PRUEBA_FS_1
- PRUEBA_FS_2
- PRUEBA_SALVATIONS_EDGE

To test memory you don't have to run an .sh file but run the processes specified in the document through API requests.
## Authors

A project of such magnitude, in such a short time could not have been completed just by myself. I am lucky to say that I had an amazing team working by my side. Check their profiles to see what they are working on! 

- [CeciC24](https://github.com/CeciC24)
- [victoriasolyedid](https://github.com/victoriasolyedid)
- [Mili-rulio](https://github.com/Mili-rulio)
- [MartuRoldan](https://github.com/MartuRoldan)
