package main

import (
	"fmt"
	"net/http"
)

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.Header().Set("Location", "/")
			w.WriteHeader(http.StatusMovedPermanently)
			return
		}

		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
<title>Mandelbrot Map</title>
<script src="https://maps.googleapis.com/maps/api/js?sensor=false"></script>
<script>
var mandelbrotTypeOptions = {
	getTileUrl: function(coord, zoom) {
		return '/mandelbrot/' + zoom + '/' + coord.x + '/' + coord.y + '.png';
	},
	tileSize: new google.maps.Size(256, 256),
	maxZoom: (1<<8) - 1,
	minZoom: 0,
	name: 'Mandelbrot'
};
	
var mandelbrotMapType = new google.maps.ImageMapType(mandelbrotTypeOptions);
		
function initialize() {
	var mapOptions = {
		center: new google.maps.LatLng(0, 0),
		zoom: 1,
		streetViewControl: false,
		mapTypeControlOptions: {
			mapTypeIds: ['mandelbrot']
		}
	};
	
	var map = new google.maps.Map(document.getElementById('map_canvas'), mapOptions);
	map.mapTypes.set('mandelbrot', mandelbrotMapType);
	map.setMapTypeId('mandelbrot');
}
</script>
<style>
#map_canvas {
	position: absolute !important;
	left: 0 !important;
	right: 0 !important;
	top: 0 !important;
	bottom: 0 !important;
}
</style>
</head>
<body onload="initialize()">
<div id="map_canvas"></div>
</body>
</html>`)
	})
}
