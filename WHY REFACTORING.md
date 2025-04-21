# Why I've refactored the API?

- Replace package level state with data  
  The shared state is an obstacle to using the functionality in parallelized tests.
- Uniform the configuration mechanism  
  It used both "flags" (passed to the factory function) and the package level variables.  
  Package level variables are now gone; only a factory function accepts options.
- Replace use of "equality" terms with "comparison"  
  `Equal` is conventionally returns true/false, depending on things being equivalent (or identical); the original implementation was using it to return a collection of
  differences. So I have renamed it as the `Compare` method, returning a `Diff`.