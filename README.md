# KWAS

K-mer-wide association study.

## Required software

- Python 3.6+ with installed libraries:
  - sklearn
  - matplotlib
  - logomaker
- [KMC 3](https://github.com/refresh-bio/KMC/releases)
- [Bowtie 2](https://github.com/BenLangmead/bowtie2/releases)
- [Diamond](https://github.com/bbuchfink/diamond/releases)
- GCC (on Windows use [msys2](https://www.msys2.org/);
  working on removing this requirement)

## Usage

This demonstration uses bash syntax,
but the pipeline can run on any platform.
Compiled binaries are available in
[releases](https://github.com/fluhus/kwas/releases).

### 1. K-mer extraction

#### 1.1. Create a file list

Create a text file containing the paths to the input fastq files,
one per line.
For this demonstration we will assume it is named `files.txt`.

#### 1.2. Count k-mers

Assuming there are `n` extraction jobs and this is job `i`:

```bash
count -f files.txt -p $i -np $n -o counts_part_$i.gz
```

#### 1.3. Merge counts

```bash
merge -t count -i "counts_part_*.gz" -o counts_all.gz
```

#### 1.4. Filter counts

If `m` is the minimal sample count for testing a k-mer:

```bash
filter -i counts_all.gz -o counts_filtered.gz -n $m
```

#### 1.5. Extract k-mer presence (HAS files) for abundant k-mers

Assuming there are `n` extraction jobs and this is job `i`:

```bash
has -f files.txt -i counts_filtered.gz -p $i -np $n -o has_part_$i.gz
```

#### 1.6. Merge k-mer presence

```bash
merge -t has -i "has_part_*.gz" -o has_all.gz
```

#### 1.7. Split by minimizer

Using minimizers of length `z` (the paper uses `z=9`):

```bash
split -i has_all.gz -o "has_part_*.gz" -k $z
```

#### 1.8. Cluster minimizers

```bash
mnzfiles -i "has_part_*.gz" -o mnz_files.txt
# For each file f in mnz_files.txt:
mnzgraph -i $f -o "$(basename $f .gz)_centers.gz" -t $num_threads
```

### 2. Population structure

#### 2.1. Subsample k-mers and samples

Using `1/n` k-mers and `1/s` samples for population structure:

```bash
smpkmers -r $n -i "has_part_*_centers.gz" -o has_popstr_tmp.gz
smpkmers -s $s -i has_popstr_tmp.gz -o has_popstr.gz
hastojson -i has_popstr.gz -o has_popstr.json
```

#### 2.2. Create projection "matrix"

Edit `popstr/popstr.py` constants with `has_popstr.json` as the input.
Run it to create the projection information.

#### 2.3. Extract population structure covariates

Assuming the result of the previous stage is `components.json`:

```bash
# For each file in files.txt:
projectpopstr -c components.json -i file_123.fq -o file_123.popstr.json
```

### 3. KWAS

#### 3.1. Create covariates table

Create an `.h5` file where the rows are samples and columns are covariates.
Add the population structure projections to this matrix.

#### 3.2. Create sample mapping

Create a JSON file that contains a map from sample file to sample name
in the covariates table.
Edit `kwas/loader.py` and set `_SAMPLE_MAPPING_FILE` to that JSON file.

#### 3.3. Run KWAS

For each file `f` in `has_part_*_centers.gz`:

```bash
python kwas/kwas.py -i $f -o $f.kwas.csv -c covariates.h5 -s files.txt
```

#### 3.4. Collect significant associations

Assuming significance threshold `p`, for each KWAS output file:

```bash
postkwas -i $f -o $f.significant -p $p
postkwas -i $f -o $f.nonsignificant -p $p -n
```

#### 3.5. Extract lists of significant and nonsignificant k-mers

For each file `f` from the previous stage:

```bash
cut -d, -f1 $f.significant | tail -n+2 > kmers.significant.txt
cut -d, -f1 $f.nonsignificant | tail -n+2 > kmers.nonsignificant.txt
```

Then concatenate all the significant ones into one file,
and the nonsignificant ones into another file.

### 4. Enrichment analysis

#### 4.1. Map to a reference

For each sample, use Bowtie or Diamond to map it to a reference.

#### 4.2. Extract reference counts

```bash
smfq -s file.sam -k kmers.significant.txt -o file.significant.json
smfq -s file.sam -k kmers.nonsignificant.txt -o file.nonsignificant.json
```

#### 4.3. Merge reference counts

```bash
smfqmerge -i "file*.significant.json" -o merged.significant.json
smfqmerge -i "file*.nonsignificant.json" -o merged.nonsignificant.json
```

#### 4.4. Run hypergeometric tests

```bash
python smgqhg -s merged.significant.json -n merged.nonsignificant.json -o rnames.json
```

Creates a JSON object where `sig` contains a list of reference names that were
enriched with significant associations,
and `found` contains all the reference names that were encountered.
