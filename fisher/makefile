# From any directory:
# make -C this_directory

.PHONY: so install uninstall

so: myfisher.cc
	python setup.py build_ext --inplace
	mv myfisher.*.so myfisher.so
	rm -fr build

install:
	python setup.py install --user
	rm -fr build dist myfisher.egg-info

uninstall:
	pip uninstall myfisher
