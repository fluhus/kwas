import json
import os
from argparse import ArgumentParser
from collections import defaultdict
from itertools import islice
from os.path import join
from subprocess import PIPE, Popen
from typing import Iterable

import numpy as np
import pandas as pd
from matplotlib import pyplot as plt
from numpy.linalg import norm
from scipy.sparse import load_npz, save_npz
from sklearn.decomposition import PCA
from sklearn.manifold import MDS
from sklearn.preprocessing import StandardScaler

plt.style.use('ggplot')

# TODO(amit): Make PCAs run on entire matrix and use disjoint cells.
# TODO(amit): Put main code in a function.


def try_setproctitle():
    """Sets the process name if the setproctitle library is available."""
    try:
        from setproctitle import setproctitle
    except ModuleNotFoundError:
        return
    setproctitle('popstr')


def save_sparse_df(df: pd.DataFrame, path: str):
    """Saves a sparse dataframe to a file."""
    save_npz(path + '.npz', df.sparse.to_coo())
    json.dump([df.index.tolist(), df.columns.tolist()],
              open(path + '.json', 'wt'))


def load_sparse_df(path: str) -> pd.DataFrame:
    """Loads a sparse dataframe from a file."""
    df = pd.DataFrame.sparse.from_spmatrix(load_npz(path + '.npz'))
    df.index, df.columns = json.load(open(path + '.json'))
    return df


def read_has() -> Iterable[dict]:
    """Reads a HAS file and returns an iterable of kmers."""
    p = Popen(['hastojson', '-i', infile], text=True, stdout=PIPE)
    return (json.loads(line) for line in p.stdout)


def read_df(short: int = None) -> pd.DataFrame:
    """Reads a samples*kmers dataframe from a HAS file."""
    fname = join(outdir, 'popstr_df')
    if os.path.exists(fname + '.npz'):
        print('Loading npz')
        return load_sparse_df(fname)
    raw = read_has()
    if short:
        raw = islice(raw, short)

    # Separate kmers from samples.
    kmers = []
    dicts = ([
        kmers.append(obj['kmer']),
        {x: 1
         for x in obj['samples']},
    ][1] for obj in raw)

    print('Building dataframe')
    df = pd.DataFrame(dicts, dtype=float)
    df.index = kmers
    print('Fixing NAs')
    df.replace(np.nan, 0, inplace=True)
    print('Converting type')
    df = df.transpose().astype('Sparse[int8]')
    print('Saving')
    save_sparse_df(df, fname)
    return df


def random_rows_cols(a, r, c):
    """Returns indexes of random selection of r rows and c columns."""
    rows = subset(a.shape[0], r)
    cols = subset(a.shape[1], c)
    return rows, cols


def shuffle_rows_cols(mm):
    """Returns mm with its rows and columns shuffled."""
    rows, cols = random_rows_cols(mm, mm.shape[0], mm.shape[1])
    return mm[rows][:, cols]


def mini_pca(a, rows, cols, b=None, n=2):
    """Projects a (or b) of a PCA space created with the given row and columns
    indexes."""
    if b is None:
        b = a
    return StandardScaler().fit_transform(
        PCA(n).fit(a[rows][:, cols]).transform(b[:, cols]))


def my_mds(x, n=2):
    """Runs MDS on the given matrix."""
    return MDS(n).fit_transform(x, init=PCA(n).fit_transform(x))


def cosine(a, b):
    """Returns the cosine similarity between a and b."""
    return np.dot(a, b) / (norm(a) * norm(b))


def my_mds_cosine(x, n=2):
    """Runs MDS using cosine dissimilarity."""
    dists = np.array([[1 - cosine(a, b) for a in x] for b in x])
    return MDS(n, dissimilarity='precomputed').fit_transform(
        dists,
        init=PCA(n).fit_transform(x),
    )


def subset(a, b):
    """Returns a random subset of a of size b."""
    return np.random.choice(a, b, replace=False)


def pca_distances(rows, cols, drows, n=2):
    """Returns an array of pairwise distances of elements in m after PCA."""
    mini = mini_pca(m, rows, cols, b=m2, n=n)
    mini = mini[drows]
    d = np.array(
        [norm(u - v) for i, u in enumerate(mini) for v in mini[i + 1:]])
    d /= norm(d) * 2**0.5
    return d


def groupby(items, key_func, value_func):
    """"Groups items into lists by their key function."""
    result = defaultdict(list)
    for x in items:
        result[key_func(x)].append(value_func(x))
    return result


def create_final_pca(n):
    """Creates the final projection matrix for use in downstream analysis."""
    print('Creating final PCA')
    pca = PCA(n).fit(df.values)

    comp = pca.components_
    evr = pca.explained_variance_ratio_
    print('Explained variance:', evr, 'Sum:', evr.sum())
    del pca

    fout = join(outdir, 'popstr.json')
    print('Writing to ' + fout)
    header = df.columns.tolist()
    with open(fout, 'wt') as f:
        for row in comp:
            json.dump({h: v for h, v in zip(header, row)}, f)
            f.write('\n')

    evr_file = fout[:-5] + '.explnvrnc.json'
    print('Writing explained variance to:', evr_file)
    json.dump(evr.tolist(), open(evr_file, 'wt'))


def plot_subsample_projections(steps):
    plt.figure(dpi=150, figsize=(15, 10))
    for i, a in enumerate(steps):
        plt.subplot(231 + i)
        print('PCA')
        randr, randc = random_rows_cols(m, rr // a, cc // a)
        mini = mini_pca(m, randr, randc, b=m2)
        plt.scatter(mini[:, 0], mini[:, 1], alpha=0.3)
        plt.xlabel(f'{rr//a} samples, {cc//a} k-mers')
    plt.subplot(232)
    plt.title('Population Structure PCA for\nDifferent Data Subsamples')
    plt.tight_layout()
    plt.savefig(join(outdir, 'popstr_pca.png'))
    plt.close()


def plot_distances_mds(rats):
    plt.figure(dpi=150)
    max_iter = 2
    dists = []
    rrows = subset(rr, rr // 10)
    groups = []
    labels = []
    for rat in rats:
        for a in range(min(rat, max_iter)):
            for b in range(min(rat, max_iter)):
                rrange = np.arange(round(rr / rat * a),
                                   round(rr / rat * (a + 1)))
                crange = np.arange(round(cc / rat * b),
                                   round(cc / rat * (b + 1)))
                print(f'PCA {rat} {a} {b}')
                dists.append(pca_distances(rrange, crange, rrows, n=2))
                groups.append(f'{rr//rat} samples, {cc//rat} k-mers')
                labels.append(f'{rat},{a},{b}')
    dists = np.array(dists)
    mds = my_mds_cosine(dists)
    for k, v in groupby(zip(mds, groups), lambda x: x[1],
                        lambda x: x[0]).items():
        arr = np.array(v)
        plt.scatter(arr[:, 0], arr[:, 1], alpha=0.5, label=k)
    plt.legend()
    plt.title('PCoA of Distance Vectors for\nDifferent Subsample Sizes')
    plt.tight_layout()
    plt.savefig(join(outdir, 'popstr_mds.png'))
    plt.close()


try_setproctitle()

parser = ArgumentParser()
parser.add_argument('-o', type=str, help='Output directory', default='.')
parser.add_argument('-i', type=str, help='Input HAS file', required=True)
args = parser.parse_args()

infile = args.i
outdir = args.o

os.makedirs(outdir)

df = read_df()
m = df.values
print('Shuffling')
m = shuffle_rows_cols(m)
print(m.shape, m.dtype)
rr, cc = m.shape

# Divide samples to 2, one half for calculating PCA and one for testing.
m2 = m[rr // 2:]
m = m[:rr // 2]
rr, cc = m.shape

plot_subsample_projections([300, 100, 30, 10, 3, 1])
plot_distances_mds([300, 100, 30, 10, 3, 1])
create_final_pca(10)
