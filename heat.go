package main

import "fmt"
import "os"
import "path"
import "log"
import "runtime/pprof"
import "github.com/codegangsta/cli"

func p( t string) {
  fmt.Println("> ", t)
}
// Profiling works by starting the program with:
//    ./heat --verbose --cpuprofile "prof.prof" on aseg.mgz --t0 42 --t1 50 \
//           --t1 43 --t1 77 --t1 63 --t1 44 --s 41 --s 51 --s 52 --s 49 \
//           --s 62 --s 53 --s 54 --s 58 --iterations 200 --distancefield 8
// followed by:
//    go tool pprof heat prof.prof 
//    list simulate
//    web

// Example calls:
// ./heat on aseg.mgz --temp0 50 --temp0 11 --temp1 42 --temp1 3 --simulate 41 --simulate 2
// ./heat on aseg.mgz --temp0 50 --temp0 77 --temp0 43 --temp0 38 --temp0 4 --temp0 11 --temp1 42 --temp1 3 --simulate 41 --simulate 2 --simulate 13 --simulate 12 --simulate 29 --simulate 30 --simulate 27 --simulate 10
// Single hemisphere (FreeSurfer label):
// ./heat --verbose on aseg.mgz --t0 42 --t1 50 --t1 43 --t1 77 --t1 63 --t1 44 --s 41 --s 51 --s 52 --s 49 --s 62 --s 53 --s 54 --s 58 --iterations "200"

func main() {

     app := cli.NewApp()
     app.Name    = "heat"
     app.Usage   = "Solving the heat equation on a 3D grid"
     app.Version = "0.0.1"
     app.Author  = "Hauke Bartsch"
     app.Email   = "HaukeBartsch@gmail.com"
     app.Flags = []cli.Flag {
       cli.BoolFlag {
         Name:  "verbose",
         Usage: "Generate verbose output with intermediate files.",
       },
       cli.StringFlag {
         Name: "cpuprofile",
         Value: "",
         Usage: "Specify a file to store profiling information",
       },
     }

     app.Commands = []cli.Command {
       {
         Name: "on",
         ShortName: "on",
         Usage: "Compute distance based sub-divisions of regions of interest by solving the heat equation.",
         Description: "Uses a label field (mgz-format) to solve the heat equation given a set of labels.\n\n" +
                      "   The --temp1 and --temp0 switches will fix the temperatures for labels in\n" +
                      "   the volume to low and high. The --simulate switch identifies label for\n" +
                      "   which the heat distribution will be simulated.\n\n" +
                      "   Most likely you will want to specify the --label <N> option to generate\n" +
                      "   individual label based on the calculated distances. The segments are created\n" +
                      "   so that each region has approximately the same number of voxel. This operation\n" +
                      "   can only succeed if the simulation resulted in a suffient number of voxel\n" +
                      "   for each range of temperature values.\n\n" +
                      "   Example:\n" + 
                      "     heat --verbose on aseg.mgz --t0 1 --t0 2 --t1 4 --s 3 -s 5 --label 3" ,
         Flags: []cli.Flag{
           cli.IntSliceFlag {
             Name: "temp0,t0",
             Value: &cli.IntSlice{},
             Usage: "Identify segments which have a low temperature. Can be specified more than once.",
           },
           cli.IntSliceFlag {
             Name: "temp1,t1",
             Value: &cli.IntSlice{},
             Usage: "Segments which has a high temperature",
           },
           cli.IntSliceFlag {
             Name: "simulate,s",
             Value: &cli.IntSlice{},
             Usage: "Segments for which the heat equation will be solved",
           },
           cli.Float64Flag {
             Name: "stepsize",
             Value: 0.12,
             Usage: "Simulation step size, should be small enough to not get Inf values",
           },
           cli.IntFlag {
             Name: "iterations",
             Value: 100,
             Usage: "Number of iterations performed",
           },
           cli.IntFlag {
             Name: "label",
             Value: 3,
             Usage: "Create a distance field with N separations for the simulated segments",
           },
           cli.BoolFlag {
             Name: "showAllTemps",
             Usage: "Show all voxel temperatures, not just the simulated subset",
           },
           cli.BoolFlag {
             Name: "gradient",
             Usage: "Create the gradient of the temperature field (nframes=3)",
           },
         },
         Action: func(c *cli.Context) {
           if len(c.Args()) < 1 {
             fmt.Printf("  Error: Specify an input label field as mgh file\n\n")
           } else {
             verbose     := c.GlobalBool("verbose")
             if (verbose) {
               p("verbose on")
               p("run heat equation")
             }

             if c.GlobalIsSet("cpuprofile") {
               fn := c.GlobalString("cpuprofile")
               f, err := os.Create(fn)
               if err != nil {
                 log.Fatal(err)
               }
               pprof.StartCPUProfile(f)
               defer pprof.StopCPUProfile()
             }


             labels, header := readMGH( c.Args()[0], verbose )
             
             temp0 := c.IntSlice("temp0")
             temp1 := c.IntSlice("temp1")
             sim   := c.IntSlice("simulate")
             omega := c.Float64("stepsize")
             iterations := c.Int("iterations")
             
             field := simulate(labels, temp0, temp1, sim, float32(omega), iterations, c.Bool("showAllTemps"), verbose)
 
             d, f  := path.Split(c.Args()[0])
             if c.IsSet("label") {
               // save a distance field version of the data (from low to high temperature in uniform intervals
               label := computeDistanceField(field, labels, sim, c.Int("label"), verbose)
               fn    := path.Join(d, f[0:len(f)-len(path.Ext(f))] + "_label.mgz")
               saveMGHuint8(label, fn, header, verbose)           
             }
             
             if c.IsSet("gradient") {
               // save the gradient of the temperature field
               gradient := computeGradientField(field, labels, sim)
               fn    := path.Join(d, f[0:len(f)-len(path.Ext(f))] + "_gradient.mgz")
               saveMGHgradient(gradient, fn, header, verbose)
             }
             
             fn    := path.Join(d, f[0:len(f)-len(path.Ext(f))] + "_temperatur.mgz")
             saveMGH(field, fn, header, verbose)
           }
         },
       },
     }
     app.Run(os.Args)
}
