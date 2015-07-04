Segmentation using geodesic distances
======================================

This is a small project that explores if the solution to the heat equation can be
used to segment white matter structures as defined by FreeSurfer's aseg.mgz based 
on shape alone (FreeSurfer aseg.mgz).

A structure that is bounded by two other non-intersecting structures can be subdivided
by this module into a number of label that form shells from the first boundary segment
to the second boundary segment. For example, the white matter of the human brain extends
from the lateral ventricles to the cortical gray matter. Using the created label field
one can identify for example lesions that are close to either the ventricles, the cortical
gray matter or at intermediate distances between both structures.

