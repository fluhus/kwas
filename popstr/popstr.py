import json
import os
from argparse import ArgumentParser
from collections import defaultdict
from ctypes import CDLL, CFUNCTYPE, POINTER, c_char_p, c_int64, c_uint8
from os.path import join

import numpy as np
import pandas as pd
from matplotlib import pyplot as plt
from numpy.linalg import norm
from sklearn.decomposition import PCA
from sklearn.manifold import MDS
from sklearn.preprocessing import StandardScaler

plt.style.use('ggplot')

# TODO(amit): Put main code in a function.


def load_has_raw():
    """Calls the low-level code for loading the HAS matrix.
    Returns a numpy array of 1/0 and the kmers as strings."""
    allocfunc = CFUNCTYPE(None, c_int64, c_int64, c_int64)
    puint8 = POINTER(c_uint8)
    pstr = POINTER(c_char_p)

    load = CDLL(libfile).cLoadMatrix
    load.argtypes = [c_char_p, allocfunc, POINTER(puint8), POINTER(pstr)]

    buf: np.ndarray = None
    pbuf = (1 * puint8)(puint8())
    pkmers = (1 * pstr)(pstr())
    nkmers = 0

    @allocfunc
    def alloc(nvals, nk, k):
        """Allocates buffers for the matrix data."""
        nonlocal buf, pbuf, pkmers, nkmers
        buf = np.zeros(nvals, dtype='uint8')
        pbuf[0] = buf.ctypes.data_as(puint8)
        strs = [('\0' * (k + 1)).encode() for _ in range(nk)]
        pkmers[0] = (nk * c_char_p)(*strs)
        nkmers = nk

    load(infile.encode(), alloc, pbuf, pkmers)
    kmers = [pkmers[0][i].decode() for i in range(nkmers)]
    buf = buf.reshape([nkmers, len(buf) // nkmers])

    return buf, kmers


def load_has():
    """Returns a dataframe with HAS data."""
    buf, kmers = load_has_raw()
    df = pd.DataFrame(buf)
    df.index = kmers
    return df.transpose()


def try_setproctitle():
    """Sets the process name if the setproctitle library is available."""
    try:
        from setproctitle import setproctitle
    except ModuleNotFoundError:
        return
    setproctitle('popstr')


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


def plot_subsample_projections(steps: int):
    """Plots PCA projections of subsamples of the matrix."""
    ratios = [2**i for i in range(steps)]
    ratios.reverse()
    plt.figure(dpi=150, figsize=(15, 10))
    for i, a in enumerate(ratios):
        plt.subplot(231 + i)
        randr, randc = random_rows_cols(m, rr // a, cc // a)
        mini = mini_pca(m, randr, randc, b=m2)
        plt.scatter(mini[:, 0], mini[:, 1], alpha=0.3)
        plt.xlabel(f'{rr//a} samples, {cc//a} k-mers')
    plt.subplot(232)
    plt.title('Population Structure PCA for\nDifferent Data Subsamples')
    plt.tight_layout()
    plt.savefig(join(outdir, 'popstr_pca.png'))
    plt.close()


def plot_distances_mds(steps: int, ss_samples=False):
    """Plots distance PCoA for subsamples of different sizes."""
    plt.figure(dpi=150)
    dists = []
    rrows = subset(rr, rr)
    groups = []
    cur_rr, cur_cc = rr, cc
    for _ in range(steps):
        if ss_samples:
            # Subsample samples & kmers:
            # xo
            # oo
            blocks = [
                [[0, cur_rr // 2], [cur_cc // 2, cur_cc]],
                [[cur_rr // 2, cur_rr], [0, cur_cc // 2]],
                [[cur_rr // 2, cur_rr], [cur_cc // 2, cur_cc]],
            ]
        else:
            # Subsample only kmers:
            # xooo
            # xooo
            # xooo
            # xooo
            blocks = [[[0, cur_rr], [cur_cc * i // 4, cur_cc * (i + 1) // 4]]
                      for i in range(1, 4)]
        for block in blocks:
            rrange = np.arange(*block[0])
            crange = np.arange(*block[1])
            dists.append(pca_distances(rrange, crange, rrows, n=2))
            groups.append(f'{cur_rr} samples, {cur_cc} k-mers')
        cur_rr //= (2 if ss_samples else 1)
        cur_cc //= (2 if ss_samples else 4)
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
parser.add_argument('-s', type=str, help='Hasmat library file', required=True)
args = parser.parse_args()

infile = args.i
outdir = args.o
libfile = args.s

os.makedirs(outdir, exist_ok=True)

print('Loading data')
df = load_has()
m = df.values
nval = m.shape[0] * m.shape[1]
print(f'Shape: {m.shape} ({nval/2**20:.0f}m values)')
del nval

print('Shuffling')
m = shuffle_rows_cols(m)
rr, cc = m.shape

# Divide samples to 2, one half for calculating PCA and one for testing.
m2 = m[rr // 2:]
m = m[:rr // 2]
rr, cc = m.shape

plot_subsample_projections(6)
plot_distances_mds(5, False)
create_final_pca(10)
