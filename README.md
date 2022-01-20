# VAA_Uebung3


- The following program start a cluster with master and n worker nodes(Philisoph"$n")
- Each node first randomly selects neighbors and sends information about them to the master node.
- The master node allows the user to print all nodes with neighbors.
- The master node also allows the user to start sending rumors. and each node sends the rumors to its neighbors if it receives it for the first time.
- The user can also print information about neighboring nodes as a directed graph. it can be printed to stdr, write it to a ".dot" file, or save it as a "png" file.
- In addition, the Master node allows the user to create an undirected graph.

## How to Start the Program:

<!-- 1. It can provide a ".txt" file path as a parameter, which includes information about the cluster master and the worker node. The following example shows the format, such as the information tree to be recorded.

    **Master: 127.0.0.1:8793**\
    **node02: 127.0.0.1:8794**
    
    The program takes as the first parameter a "file" string, and the second parameter the path of the file as follows: 

    **./start file Nodes.txt**
    
    __A "Nodes.txt" file already exists in the root folder.__ -->


-   could automatically start any node number. The program accepts one integer parameter, which represents the number of nodes. The following example starts a cluster with 10 Philisoph nodes:

    **./start 10**


The program start master and worker(philisoph) nodes in a new tabs. each node shows the cycle of coordinator election and appointment negotiation. 

A menu appears on the master node 10 seconds after the cluster starts, which can help you get more information about the cluster.


## Important. 
To start the election, in Master tap, neighbors of the cluster member must be read from a "Graph.dot" file. this could be done by giving "9" to the terminal. in the root folder there is a file "mygraph.dot", which will be read by default. it could be replaced in "/Cluster/init.go file" in a global var called "file"

To modify the available time and A_Max you can also replace in this path: "/Cluster/init.go file"