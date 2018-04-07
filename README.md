# Terrarium

Some code for generating topographic contour maps. Documentation is currently lacking, so you're mostly on your own!

`cmd/test/main.go` will automatically fetch (and locally cache) [AWS Terrain tiles](https://aws.amazon.com/public-datasets/terrain/) and render a topographic isoline map as a PNG. You can specify a lat/lng bounding box or use a shapefile. See the top of [main.go](https://github.com/fogleman/terrarium/blob/master/cmd/test/main.go) to configure this.

`cmd/contours/main.go` will generate contours for a grayscale input image. Again there are some constants at the top of the file that you can configure.

## Examples

#### Colorado

![Colorado](https://i.imgur.com/ZzeDAAU.png)

#### Valles Marineris on Mars

![Valles Marineris](https://i.imgur.com/BHRpnQd.png)
