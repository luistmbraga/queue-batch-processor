# batch-request-service

This repository is a proof of concept (POC) for a proxy server that acts as an interface for another service or system with processing limitations.

This POC gathers requests from multiple sources and processes them all at once in a batch, making use of Go routines.

What happens is:
1. A request is arrives and checks if an ongoing batch is being processed;
2. If a batch is not being processed it moves further;
3. The data for processing is added to the batch data;
4. If there isn't another routine waiting for more requests, the current routine initializes a timer and waits for more requests;
5. Other requests that arrive while the timer is ongoing will add their data to the batch and wait;
6. When the timer finishes no more requests will be accepted for the batch and will be added to a waiting condition;
7. The request is made and placed on a global variable;
8. The last routine to return (counter == 0) cleans the variables and awakes the waiting threads.

A simple Jmeter test was added to test the POC. The Jmeter test is meant to be used with the given example.
