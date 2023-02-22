# cgz

Chronological dumper/filter for gzip text file without header&amp;footer

## How it works

1. Cgz explores all files in -path, it keeps all file names that follow pattern {epoch}.json.gz, epoch behing number of seconds since 1st January 1970.
2. Sort file names using epoch in ascending order.
3. Deflate each file in stdout

## Filtering

When you use -day option you can dump file corresponding to only one day.

For example if you input -d 20230222 you will keep only files in 2023-02-22T00:00:00Z<= epoch <2023-02-22T00:00:00Z.

Internally it checks tahn epoch is in [1677024000(2023-02-22) 1677110400(2023-02-23)[ 

## Calling examples

The following example should find around 114000 files in one hour then dump and xz compress in another hour.

Full process should take about 2 hours and generate a 21MB xz file from an HDD.

### Dump all file for 25th January 2023 into an xz file :

```bash
./cgz -path /mnt/raid0/data/magiline/boxes/ -day 20230125 | xz --best >20230125_boxes.xz
```

It will promp progression every 10 seconds.

### Same as previous but in background and logging progression/errors :

```bash
nohup ./cgz -path /mnt/raid0/data/magiline/boxes/ -day 20230125 2>20230125_boxes.log | xz --best >20230125_boxes.xz &
```
