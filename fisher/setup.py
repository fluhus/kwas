# Place the .pyd file where the library's python file would be.

from distutils.core import setup, Extension

setup(name='myfisher',
      version='1.0',
      ext_modules=[
          Extension(
              'myfisher',
              ['myfisher.cc'],
              extra_compile_args=['-std=c++20', '-Wall'],
          )
      ])
