"""Fisher test for samfastq output."""
import gzip
import json
import math

import myfisher
import numpy as np
from matplotlib import pyplot as plt

NAMES_FILE = 'TODO'
COUNTS_FILE = 'TODO'

BATCH = 'dmnd'
MERGE_SIG_FILE = f'merge.{BATCH}.sig.gz'
MERGE_NSIG_FILE = f'merge.{BATCH}.nsig.gz'
OUTPUT_FILE = f'fishers.{BATCH}.json.gz'


def rename_keys(d, mapping):
    d2 = {}
    for k, v in d.items():
        d2[mapping[k]] = d2.get(mapping[k], 0) + v
    return d2


def to_gene_names(*d):
    print('Loading names')
    names = json.load(gzip.open(NAMES_FILE))
    print('Renaming')
    return tuple(rename_keys(x, names) for x in d)


def add_nice_names(dd: list[dict]):
    print('Loading names')
    names = json.load(gzip.open(NAMES_FILE))
    print('Renaming')
    for d in dd:
        d['nice'] = names[d['name']]


def load_data() -> tuple[dict[str, int], dict[str, int]]:
    print('Loading data')
    sig_counts = json.load(gzip.open(COUNTS_FILE.format('sig'), 'rt'))
    all_counts = json.load(gzip.open(COUNTS_FILE.format('all'), 'rt'))
    print('Subtracting')
    nsig_counts = {
        k: v2
        for k, v in all_counts.items() if (v2 := v - sig_counts.get(k, 0)) > 0
    }
    return sig_counts, nsig_counts


def fixinf(x):
    return x if not math.isinf(x) else -1


def load_gene_kmers(f):
    j = (json.loads(row) for row in gzip.open(f))
    d = {}
    s = set()
    for x in j:
        d[x['Gene']] = len(x['Kmers'])
        s.update(x['Kmers'])
    return d, len(s)


def species_fisher():
    print('Loading sig counts: ' + MERGE_SIG_FILE)
    sigcounts, sigsum = load_gene_kmers(MERGE_SIG_FILE)
    print(len(sigcounts))
    print('Loading nsig counts: ' + MERGE_NSIG_FILE)
    nsigcounts, nsigsum = load_gene_kmers(MERGE_NSIG_FILE)
    print(len(nsigcounts))

    PLOT_A_DIST = False
    if PLOT_A_DIST:
        plt.style.use('ggplot')
        prc = np.percentile(list(sigcounts.values()), list(range(101)))
        plt.scatter(list(range(101)), prc, alpha=0.5)
        plt.yscale('log')
        plt.xlabel('Percentile')
        plt.ylabel('Significant kmer count')
        plt.show()
        return

    print('Fishing')
    tests = []
    for k in sigcounts:
        a = sigcounts[k]
        b = sigsum - a
        c = nsigcounts.get(k, 0)
        d = nsigsum - c
        tests.append((k, *myfisher.fisher(a, b, c, d, 'greater')))

    tests = [t for t in tests if t[2] < 0.05 / len(tests)]
    print(len(tests), 'significant')
    # return
    tests = sorted(tests, key=lambda x: -x[1])
    print(tests[:3])

    sig = [x[0] for x in tests]
    found = list(set(sigcounts) | set(nsigcounts))
    out = {'sig': sig, 'found': found}

    print('Writing to: ' + OUTPUT_FILE)
    json.dump(out, gzip.open(OUTPUT_FILE, 'wt'), indent=2)


species_fisher()
