# cgz
Chronological dumper/filter for gzip text file without header&amp;footer

## How it works

### 1st step files selection

Cgz explores all files in -path, it keeps all file names taht follow pattern {epoch}.json.gz, epoch behing number of seconds since 1st January 1970.

