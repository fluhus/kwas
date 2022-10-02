"""Loader for kmer presence/absence ("has")."""
from ctypes import CDLL, POINTER, Structure, c_bool, c_char_p, c_int8, \
    c_size_t, c_uint64, c_void_p, cast
from datetime import datetime
import json
from typing import List

import numpy as np
import pandas as pd

# The underlying low-level implementation of the loader.
_LIBRARY_PATH = 'kmerloader.so'
# JSON file with sample-to-participant mapping.
_SAMPLE_MAPPING_FILE = 'TODO'


class Error(Structure):
    """Represents an error that might return from the low-level code.
    If no error, err will be none."""
    _fields_ = [('err', c_char_p)]

    def __del__(self):
        if self.err is not None:
            _del_error(self)

    def raise_if_error(self):
        if self.err is not None:
            raise IOError(self.err.decode())


class CLoader(Structure):
    """Contains a handle to the low-level loader."""
    _fields_ = [('err', Error), ('handle', c_size_t)]

    def __del__(self):
        if self.handle:
            _close(self)


class Strings(Structure):
    """An array of C-strings."""
    _fields_ = [('strs', POINTER(c_char_p))]

    def __del__(self):
        _del_strings(self)

    def strings(self) -> List[str]:
        """Returns the value as python strings."""
        result = []
        for s in self.strs:
            if s is None:
                break
            result.append(s.decode())
        return result


class KmerHas(Structure):
    """Holds the samples that contain a specific kmer."""
    _fields_ = [
        ('err', Error),
        ('kmer', c_char_p),
        ('has', POINTER(c_bool)),
        ('n', c_uint64),
    ]

    def __del__(self):
        _del_has(self)


class KmerHasLoader():
    """Loads kmer presence/absence ("has") information."""
    def __init__(self, samples_fname: str, has_fname: str):
        super().__init__()
        self._ld = _open(samples_fname.encode(), has_fname.encode())

        # From sample to registration code
        sm = json.load(open(_SAMPLE_MAPPING_FILE))
        samples = _get_samples(self._ld).strings()
        samples = [sm[x] for x in samples]

        self._df = pd.DataFrame({
            'RegistrationCode': samples,
            'Has': [False] * len(samples),
        }).set_index('RegistrationCode')

    def __iter__(self):
        return self

    def __next__(self):
        data = self.get_data()
        if data is None:
            raise StopIteration
        return data

    @property
    def _data_column_names(self):
        return ['Has']

    @property
    def _data_index_names(self):
        return ['RegistrationCode']

    def _load_data(self):
        """No-op. Loading is done in get_data."""
        pass

    def get_data(self):
        """Returns information for a single kmer. Subsequent calls return subsequent
        kmers. The kmer sequence is available at get_data().df.attrs['kmer']."""
        raw = _next(self._ld)

        # n=0 means no more data (EOF).
        if raw.n == 0:
            return None

        # Assigning to the DF with a numpy array because it's faster.
        addr = cast(raw.has, c_void_p).value
        sized = (c_int8 * raw.n).from_address(addr)

        self._df['Has'] = np.array(sized)
        self._df.attrs['kmer'] = raw.kmer.decode()
        return self._df.copy()


def _check_err(result, unused_func, unused_arguments):
    """Error checker for low-level functions that return an error."""
    result.raise_if_error()
    return result


def _check_err_field(result, unused_func, unused_arguments):
    """Error checker for low-level functions that return an error field."""
    result.err.raise_if_error()
    return result


# Hooks for calling the low-level code.
_lib = CDLL(_LIBRARY_PATH)

_open = _lib.open
_open.argtypes = [c_char_p, c_char_p]
_open.restype = CLoader
_open.errcheck = _check_err_field

_close = _lib.close
_close.argtypes = [CLoader]
_close.restype = Error
_close.errcheck = _check_err

_get_samples = _lib.getSamples
_get_samples.argtypes = [CLoader]
_get_samples.restype = Strings

_next = _lib.next
_next.argtypes = [CLoader]
_next.restype = KmerHas
_next.errcheck = _check_err_field

_del_error = _lib.delError
_del_error.argtypes = [Error]

_del_strings = _lib.delStrings
_del_strings.argtypes = [Strings]

_del_has = _lib.delHas
_del_has.argtypes = [KmerHas]


def _sanity_check():
    """A simple check to verify that the loader works."""
    print('Creating loader')
    t = datetime.now()
    ld = KmerHasLoader(
        '/home/amitmit/Data/genie/LabData/Analyses/amitmit/kmers/samples_d2.txt',
        '/home/amitmit/Data/mb/amitmit/kmers/maf/1.has.gz')
    print(datetime.now() - t)

    df = ld.get_data().df
    print(df.attrs)
    print(df.sum())
    df = ld.get_data().df
    print(df.attrs)
    print(df.sum())

    print('Testing load speed')
    n = 1000
    t = datetime.now()
    for _ in range(n):
        ld.get_data()
    d = datetime.now() - t
    print(f'{d} ({d/n} per item)')


if __name__ == '__main__':
    _sanity_check()
