"""Loads kmer presence (HAS) files."""
import subprocess as sp
import json


class HasLoader:
    """An iterator on HAS files."""
    p: sp.Popen

    def __init__(self, exe: str, has: str):
        """Creates a new loader. Arguments are the paths to the hastojson
        executable and the input HAS file."""
        self.p = sp.Popen([exe, '-i', has],
                          bufsize=1,
                          stdout=sp.PIPE,
                          text=True)

    def __del__(self):
        if hasattr(self, 'p') and self.p.poll() is None:
            self.p.kill()

    def __iter__(self):
        return self

    def __next__(self):
        line = self.p.stdout.readline()
        if not line:
            raise StopIteration
        return json.loads(line)
