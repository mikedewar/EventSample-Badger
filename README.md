# EventSample-Badger

Generate a set of events by sampling from a poisson process from each edge in
an edgelist stored in a badgerDB.

This iterates through every edge in the table, and pulls out events from each
edge. It then writes those events to another (much much larger) table keyed by
time. The idea being then we can then run through the event table and get
a time ordered set of events to put in kafka (or wherever) for downstream
testing. 

Choose λ as your rate for every edge.

# TODO

* needs to be keyed by time+uuid else we'll get loads of clashes.
* specify a distribution for lambda and sample for each edge, rather than
    having one global rate.
