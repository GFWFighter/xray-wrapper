apple:
	@gomobile bind -o ./build/XRay.xcframework -target=ios,macos,iossimulator -ldflags="-s -w" -v ./
ios:
	@gomobile bind -o ./build/ios/XRay.xcframework -target=ios,iossimulator -ldflags="-s -w" -v ./
macos:
	@gomobile bind -o ./build/macos/XRay.xcframework -target=macos -ldflags="-s -w" -v ./
	
