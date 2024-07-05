.PHONY: default bundle

# Default to help
default: bundle

# bundle, @see github.com/bagaking/file_bundle
bundle:
	$(MAKE) -C bundle -f Makefile clean
	$(MAKE) -f bundle/Makefile
	#file_bundle -v -i ./bundle/_.file_bundle_rc -o ./bundle/_.bundle.txt
