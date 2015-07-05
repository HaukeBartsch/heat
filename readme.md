Segmentation using geodesic distances
======================================

This is a small project that explores if the solution to the heat equation can be
used to segment white matter structures as defined by FreeSurfer's aseg.mgz based 
on shape alone (FreeSurfer aseg.mgz). The program reads in an mgz label
file (unsigned char, type == 0) and three regions of interest as defined by their label
number. It produces a floating point mgz file with simulated temperature values
for each voxel for which the simulation was run. It can also export regions of
interest that separate the simulated region into discreet regions at temperature
iso-lines. 

A structure that is bounded by two other non-intersecting structures can be subdivided
by this module into a number of label that form shells from the first boundary segment
to the second boundary segment. Topologically this corresponds to a sphere with a sphere 
inside which has a third sphere inside. For example, the white matter of the human brain extends
from the lateral ventricles to the cortical gray matter. Using the created label field
one can identify for example lesions that are close to either the ventricles, the cortical
gray matter or, at intermediate distances between both structures.

More information about the rationale for this program can be found at http://haukebartsch.github.io/heat/.

Download binary:

  * MacOS command line executable
    https://github.com/HaukeBartsch/heat/raw/master/binary/MacOS/heat

  * Linux64 command line executable
    https://github.com/HaukeBartsch/heat/raw/master/binary/Linux64/heat

```
NAME:
   heat - Solving the heat equation on a 3D grid

USAGE:
   heat [global options] command [command options] [arguments...]

VERSION:
   0.0.1

AUTHOR:
  Hauke Bartsch - <HaukeBartsch@gmail.com>

COMMANDS:
   on, on	Compute distance based sub-divisions of regions of interest by solving the heat equation.
   help, h	Shows a list of commands or help for one command
   
GLOBAL OPTIONS:
   --verbose		Generate verbose output with intermediate files.
   --cpuprofile 	Specify a file to store profiling information
   --help, -h		show help
   --version, -v	print the version
```

The 'on' sub-command allows to specify a number of additional options:

```
NAME:
   on - Compute distance based sub-divisions of regions of interest by solving the heat equation.

USAGE:
   command on [command options] [arguments...]

DESCRIPTION:
   Uses a label field (mgz-format) to solve the heat equation given a set of labels.

   The --temp1 and --temp0 switches will fix the temperatures for labels in
   the volume to low and high. The --simulate switch identifies label for
   which the heat distribution will be simulated.

   Most likely you will want to specify the --label <N> option to generate
   individual label based on the calculated distances. The segments are created
   so that each region has approximately the same number of voxel. This operation
   can only succeed if the simulation resulted in a suffient number of voxel
   for each range of temperature values.

   Example:
     heat --verbose on aseg.mgz --t0 1 --t0 2 --t1 4 --s 3 -s 5 --label 3

OPTIONS:
   --temp0, --t0 [--temp0 option --temp0 option]	Identify segments which have a low temperature. Can be specified more than once.
   --temp1, --t1 [--temp1 option --temp1 option]	Segments which has a high temperature
   --simulate, -s [--simulate option --simulate option]	Segments for which the heat equation will be solved
   --stepsize "0.12"					Simulation step size, should be small enough to not get Inf values
   --iterations "100"					Number of iterations performed
   --label "3"						Create a distance field with N separations for the simulated segments
   --showAllTemps					Show all voxel temperatures, not just the simulated subset
   --gradient						Create the gradient of the temperature field (nframes=3)
```
