Exercise 1 - Theory questions
-----------------------------

### Concepts

What is the difference between *concurrency* and *parallelism*?
> Concurrency is managing multiple tasks at the "same" time, but in an overlapping manner (the system switches between them), while parallelism actually executes the multiple tasks simultaneous by using multiple resources. 

What is the difference between a *race condition* and a *data race*? 
> Race condition happens when the correctness of the code is affected by the timing or order of events. A data race happens when one thread accesses a changable object while another thread is writing to it. Data race is a type of race condition. 
 
*Very* roughly - what does a *scheduler* do, and how does it do it?
> The scheduler decides the order of threads and processes. It uses algorithms to prioritize based on different factors. 
> A scheduler is a mechanism that assigns resources such as processors or network links to perform a task such as a thread or a data flow. It assigns a fixed time unit per process and cycles through them. 
Hvilken er rett???


### Engineering

Why would we use multiple threads? What kinds of problems do threads solve?
> To be able to use multiple processors at the same time, so that multiple tasks can be executed at the same time (concurrently og in parallel). It can help improve the  efficiency of a program, since it can prioritize critical tasks to avoid time delays. 

Some languages support "fibers" (sometimes called "green threads") or "coroutines"? What are they, and why would we rather use them over threads?
> Fibers are user-level implementations of threads that are scheduled by a user-level scheduler instead of the operating system. Coroutines can be paused and then resumed from where it left off. The are more lightweight and faster to create and switch between than threads. 

Does creating concurrent programs make the programmer's life easier? Harder? Maybe both?
> Both - easier because it allows cleaner code by avoiding waiting for tasks to complete, but harder because it needs synchronization, and also has non-deterministic behaviour, making it harder to predict and reproduce issues. 

What do you think is best - *shared variables* or *message passing*?
> *Your answer here*
Message passing is safer, since it avoids issues with race conditions and synchronization by not having shared state. But, shared variables are better for applications where performance is important. 


