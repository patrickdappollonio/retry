# retry
Retry is a simple package to execute an operation multiple times until a condition is met.

### Example code

```go
// Create a new retry task, which executes every 2 seconds
r := retry.New(retry.Sleep(2 * time.Second))

// Poll a httpstatus URL every n seconds
fn := func() (retry.Reason, error) {
    // Perform an http request or anything that needs to be polled every 
    // 2 seconds. Depending on the return value (retry.Reason) which could
    // be either retry.Stop or retry.Again, it'll either stop and not try again
    // or try again after 2 seconds.
    
    // This can also return an error, which comes from the function itself
    // (like something failed) or retry.ErrOSRequestedCancellation, which means
    // we received a SIGTERM / SIGKILL message to stop operations.
    
    // Query example.com...
    resp, err := http.Get("http://example.com/")
    
    // If there's an error we don't want to continue
    // so we return retry.Stop, and we also return the error
    // for further inspection.
    if err != nil {
      return retry.Stop, err
    }
    
    // Close the response body once we're finished
    defer resp.Body.Close()
    
    // Read the body
    body, err := ioutil.ReadAll(resp.Body)
    
    // If there's an error parsing the body, let's assume
    // we want to try again making the HTTP call
    if err != nil {
      return retry.Again, nil
    }
    
    // Check if the body contains the word "example", if so
    // stop the execution and don't try again
    if bytes.Contains(body, "example") {
      return retry.Stop, nil
    }
    
    // If we didn't find the word "example" in the body
    // try over and over again every 2 seconds until you get it
    return retry.Again, nil
}

// Check the overall result of the retrying process
if err := r.Do(fn); err != nil {
  if err == retry.ErrOSRequestedCancellation {
    fmt.Println("OS requested stop: SIGTERM / SIGKILL happened.")
    return
  }
  
  fmt.Println("Error found:", err.Error())
}
```
