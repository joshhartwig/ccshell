# Issues

(fixed) 1. ```cat noexist > output.txt``` pipes error to text file but should error out first, note the correct error from stdout is going into the file
2. echo to a specific path errors as echo has its own path writing that bombs on absolute paths
