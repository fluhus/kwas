"""Fisher test for samfastq output."""
import json

import myfisher
import numpy as np
from matplotlib import pyplot as plt

# TODO(amit): Make these command-line arguments.
INPUT_SIG_FILE = 'TODO'
INPUT_NSIG_FILE = 'TODO'
OUTPUT_FILE = 'TODO'


def load_gene_kmers(f):
    j = (json.loads(row) for row in open(f))
    d = {}
    s = set()
    for x in j:
        d[x['Gene']] = len(x['Kmers'])
        s.update(x['Kmers'])
    return d, len(s)


def species_fisher():
    print('Loading sig counts: ' + INPUT_SIG_FILE)
    sigcounts, sigsum = load_gene_kmers(INPUT_SIG_FILE)
    print(len(sigcounts))
    print('Loading nsig counts: ' + INPUT_NSIG_FILE)
    nsigcounts, nsigsum = load_gene_kmers(INPUT_NSIG_FILE)
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
    json.dump(out, open(OUTPUT_FILE, 'w'), indent=2)


species_fisher()
