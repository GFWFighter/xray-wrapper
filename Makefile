apple:
	@gomobile bind -o ./build/XRay.xcframework -target=ios,macos,iossimulator -ldflags="-s -w" -v ./
