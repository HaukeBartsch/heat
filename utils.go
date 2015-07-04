package main

import (
  "io"
  "net/http"
  "os"
  "log"
  "fmt"
  "path"
  "io/ioutil"
  "compress/gzip"
  "bytes"
  "math"
  "encoding/binary"
  //"image/color"
  _  "image/jpeg"
  _  "image/png"
)

type header  struct {
  version, width, height, depth, nframes, t, dof int32
  goodRASFlag int16
  vz [3]float32
  Mdc [9]float32
  Pxyz [3]float32
}

// read in mgz file - ignores all transformations
func readMGH( fn string, verbose bool ) ( [][][]uint8, header ) {

  var head header
  // find out if the file has mgz extension (read with zip file reader first)
  _, f := path.Split(fn)
  if path.Ext(f) == ".mgz" {
     // p(fmt.Sprintf("found an mgz file, unzip first"))

     fi, err := os.Open(fn)
     if err != nil {
       p(fmt.Sprintf("Error: could not open file"))
       os.Exit(-1)
     }
     defer fi.Close()
     fz, err := gzip.NewReader(fi)
     if err != nil {
       os.Exit(-1)
     }
     defer fz.Close()
     s, err := ioutil.ReadAll(fz)
     if err != nil {
       os.Exit(-1)
     }
     newfn := ".download.mgh"

     err = ioutil.WriteFile(newfn, s, 0644)
     if err != nil {
        p(fmt.Sprintf("Error: could not write temporary file"))
     }
     fn = newfn
  }

  var file *os.File
  if _, err := os.Stat(fn); err == nil { // read using direct io
     file, err = os.Open(fn)
     if err != nil {
        log.Fatal(err)
     }
  } else { // this part only works if https has a valid non-self-signed certificate
    if verbose {
      p("try to download file")
    }
    out, _ := os.Create(".download")
    defer out.Close()
    resp, err := http.Get(fn)
    if err != nil {
      log.Fatal(err)
    }
    defer resp.Body.Close()
    _, err = io.Copy(out, resp.Body)
    if err != nil {
       log.Fatal(err)
    }
    file, err = os.Open(".download")
    if err != nil {
      log.Fatal(err)
    }
  }
  defer file.Close()

  // now start reading the file (un-gzipped mgh)
  head.version = read4(file)
  head.width   = read4(file)
  head.height  = read4(file)
  head.depth   = read4(file)
  head.nframes = read4(file)
  head.t       = read4(file)
  head.dof     = read4(file)
  head.goodRASFlag = read2(file)
  if verbose {
     p(fmt.Sprintf("Input data: [version: %d, width: %d, height: %d, depth: %d, nframes: %d, type: %d, dof: %d, goodRASFlag: %d]", head.version, head.width, head.height, head.depth, head.nframes, head.t, head.dof, head.goodRASFlag))
  }
  if head.version != 1 {
    p(fmt.Sprintf("Error: this version of mgz is not supported"))
  }
 
  if head.t != 0 {
    p(fmt.Sprintf("Error: this program only support files with unsigned character encoding"))
  }
  
  if head.nframes != 1 {
    p(fmt.Sprintf("Warning: only the first frame will be read"))    
  }
  
  // create the space for the label
  var dims [3]int32
  dims[0] = head.width
  dims[1] = head.height
  dims[2] = head.depth
  // create space for unsigned char (we should check memory here)
  labels := make([][][]uint8, dims[2])
  for i := range labels {
    labels[i] = make([][]uint8, dims[1])
    for j := range labels[i] {
      labels[i][j] = make([]uint8, dims[0])
      if len(labels[i][j]) != int(dims[0]) {
        p(fmt.Sprintf("Error: could not get enough memory for label field"))
      }
    }
  }

  if head.goodRASFlag == 1 {
    // read and forget the
    head.vz[0] = read4AsFloat(file)  
    head.vz[1] = read4AsFloat(file)  
    head.vz[2] = read4AsFloat(file)
    head.Mdc[0] = read4AsFloat(file)
    head.Mdc[1] = read4AsFloat(file)
    head.Mdc[2] = read4AsFloat(file)
    head.Mdc[3] = read4AsFloat(file)
    head.Mdc[4] = read4AsFloat(file)
    head.Mdc[5] = read4AsFloat(file)
    head.Mdc[6] = read4AsFloat(file)
    head.Mdc[7] = read4AsFloat(file)
    head.Mdc[8] = read4AsFloat(file)
    head.Pxyz[0] = read4AsFloat(file)
    head.Pxyz[1] = read4AsFloat(file)
    head.Pxyz[2] = read4AsFloat(file)
  }
  
  file.Seek(284, 0)
  // now read in the data (don't need to swap because its all unsigned char)
  buf := make([]byte, int64(dims[0])*int64(dims[1])*int64(dims[2]))
  n, err := file.Read(buf)
  if err != nil {
    panic(err)
  }
  if int64(n) != int64(dims[0])*int64(dims[1])*int64(dims[2]) {
    p(fmt.Sprintf("Error: could not read all data from file"))
  }
  file.Sync()
  var count int64
  count = 0
  for k := 0; k < int(dims[2]); k++ {
    for j := 0; j < int(dims[1]); j++ {
      for i := 0; i < int(dims[0]); i++ {
        labels[k][j][i] = buf[count]
        count = count + 1
      }
    }
  }
  
  return labels, head
}

func read4AsFloat( file *os.File ) (float32) {
  
  buf4 := make([]byte, 4)  
  n, err := file.Read(buf4)
  if err != nil {
    panic(err)
  }
  if n != 4 {
    p("Error: could not read 4 bytes")
  }
  file.Sync()
  buf := bytes.NewReader(buf4)
  var val float32
  err = binary.Read(buf, binary.BigEndian, &val)
  if err != nil {
    panic(err)
  }  
  
  return val
}


func read4( file *os.File ) (int32) {
  
  buf4 := make([]byte, 4)  
  n, err := file.Read(buf4)
  if err != nil {
    panic(err)
  }
  if n != 4 {
    p("Error: could not read 4 bytes")
  }
  file.Sync()
  buf := bytes.NewReader(buf4)
  var val int32
  err = binary.Read(buf, binary.BigEndian, &val)
  if err != nil {
    panic(err)
  }  
  
  return val
}

func save4( file *os.File, val int32 ) {

  err := binary.Write(file, binary.BigEndian, val)
  if err != nil {
    panic(err)
  }
}
func save4float32( file *os.File, val float32 ) {

  err := binary.Write(file, binary.BigEndian, val)
  if err != nil {
    panic(err)
  }
}

func save2( file *os.File, val int16 ) {

  err := binary.Write(file, binary.BigEndian, val)
  if err != nil {
    panic(err)
  }
}

func read2( file *os.File ) (int16) {
  
  buf2 := make([]byte, 2)  
  n, err := file.Read(buf2)
  if err != nil {
    panic(err)
  }
  if n != 2 {
    p("Error: could not read 2 bytes")
  }
  file.Sync()
  buf := bytes.NewReader(buf2)
  var val int16
  err = binary.Read(buf, binary.BigEndian, &val)
  if err != nil {
    panic(err)
  }  
  
  return val
}


func saveMGH( field [][][]float32, fn string, head header, verbose bool) {
  if verbose {
    p(fmt.Sprintf("start writing output file %s...", fn))
  }
  // write the input field to fn, we can take the header from the parent, but we need to change the output type to float (3)
  fi, err := os.Create(fn)
  if err != nil {
     p(fmt.Sprintf("Error: could not open file %s", fn))
     os.Exit(-1)
  }
  defer fi.Close()
  
  var typ int32
  typ = 3 // save as floating point field
  save4(fi, head.version)
  save4(fi, head.width)
  save4(fi, head.height)
  save4(fi, head.depth)
  save4(fi, head.nframes)
  save4(fi, typ)
  save4(fi, head.dof)
  save2(fi, head.goodRASFlag)
  
  save4float32(fi, head.vz[0])
  save4float32(fi, head.vz[1])
  save4float32(fi, head.vz[2])
  save4float32(fi, head.Mdc[0])
  save4float32(fi, head.Mdc[1])
  save4float32(fi, head.Mdc[2])
  save4float32(fi, head.Mdc[3])
  save4float32(fi, head.Mdc[4])
  save4float32(fi, head.Mdc[5])
  save4float32(fi, head.Mdc[6])
  save4float32(fi, head.Mdc[7])
  save4float32(fi, head.Mdc[8])  
  save4float32(fi, head.Pxyz[0])
  save4float32(fi, head.Pxyz[1])
  save4float32(fi, head.Pxyz[2])
  
  // go to byte 284 (did write 82 bytes so far)
  for i := 0; i < (284-90); i++ {
    var val uint8
    val = 0
    err := binary.Write(fi, binary.BigEndian, val)
    if err != nil {
       p(fmt.Sprintf("Error: could not write bytes to output"))
    }
  }
  // now save the binary data
  for k := 0; k < int(head.depth); k++ {
    for j := 0; j < int(head.height); j++ {
        err := binary.Write(fi, binary.BigEndian, field[k][j][:])
        if err != nil {
           p(fmt.Sprintf("Error: could not write bytes to output"))          
        }
    }
  }
  fi.Sync()
  if verbose {
    p(fmt.Sprintf("writing %s done", fn))
  }
}

func saveMGHuint8( field [][][]uint8, fn string, head header, verbose bool) {
  if verbose {
    p(fmt.Sprintf("start writing output file %s...", fn))
  }
  // write the input field to fn, we can take the header from the parent, but we need to change the output type to float (3)
  fi, err := os.Create(fn)
  if err != nil {
     p(fmt.Sprintf("Error: could not open file %s", fn))
     os.Exit(-1)
  }
  defer fi.Close()
  
  var typ int32
  typ = 0 // save as floating point field
  save4(fi, head.version)
  save4(fi, head.width)
  save4(fi, head.height)
  save4(fi, head.depth)
  save4(fi, head.nframes)
  save4(fi, typ)
  save4(fi, head.dof)
  save2(fi, head.goodRASFlag)
  
  save4float32(fi, head.vz[0])
  save4float32(fi, head.vz[1])
  save4float32(fi, head.vz[2])
  save4float32(fi, head.Mdc[0])
  save4float32(fi, head.Mdc[1])
  save4float32(fi, head.Mdc[2])
  save4float32(fi, head.Mdc[3])
  save4float32(fi, head.Mdc[4])
  save4float32(fi, head.Mdc[5])
  save4float32(fi, head.Mdc[6])
  save4float32(fi, head.Mdc[7])
  save4float32(fi, head.Mdc[8])  
  save4float32(fi, head.Pxyz[0])
  save4float32(fi, head.Pxyz[1])
  save4float32(fi, head.Pxyz[2])
  
  // go to byte 284 (did write 82 bytes so far)
  for i := 0; i < (284-90); i++ {
    var val uint8
    val = 0
    err := binary.Write(fi, binary.BigEndian, val)
    if err != nil {
       p(fmt.Sprintf("Error: could not write bytes to output"))
    }
  }
  // now save the binary data
  for k := 0; k < int(head.depth); k++ {
    for j := 0; j < int(head.height); j++ {
        err := binary.Write(fi, binary.BigEndian, field[k][j][:])
        if err != nil {
           p(fmt.Sprintf("Error: could not write bytes to output"))          
        }
    }
  }
  fi.Sync()
  if verbose {
    p(fmt.Sprintf("writing %s done", fn))
  }
}


func computeDistanceField(field [][][]float32, labels [][][]uint8, simulate []int, numsegments int) ( [][][]uint8 ){
  // store end of each segment
  borders := make([]float32, numsegments-1) // keep a list of the (uniform distant) quantiles requested by the user
  for i := range borders {
    borders[i] = 1.0/float32(numsegments)*float32(i+1)
    //fmt.Printf("quantile threshold value %d is %g\n", i, borders[i])
  }

  var dims [3]int
  dims[2] = len(labels)
  dims[1] = len(labels[0])
  dims[0] = len(labels[0][0])
  df := make([][][]uint8, dims[2])
  for i := range df {
    df[i] = make([][]uint8, dims[1])
    for j := range df[i] {
      df[i][j] = make([]uint8, dims[0])
    }
  }
  
  // we will compute quantiles for the actual separations
  // we know that the temperature is between 0.01 and 0.1
  // lets define numsegments quantiles for the field values in every label of labels that is listed in simulate
  simThese := make([][][]uint8, dims[2])
  for k := range simThese {
     simThese[k] = make([][]uint8, dims[1])
     for j := range simThese[k] {
       simThese[k][j] = make([]uint8, dims[0])
       for i := range simThese[k][j] {
          simThese[k][j][i] = 0
          val := int(labels[k][j][i])
          for l := range simulate {
             if simulate[l] == val {
                simThese[k][j][i] = 1 // only export distance for these voxel
                break
             }
          }
       }
     }
  }

  // collect a histogram of heat values (use it to compute a cummulative histogram later)
  histresolution := 512
  hist := make([]int, histresolution)
  for i := range hist { hist[i] = 0 } // explicitely set to zero
  for k := range field {
    for j := range field[k] {
      for i := range field[k][j] {
        if simThese[k][j][i] == 0 {
          continue
        }
        index := int( math.Floor( float64( ( (field[k][j][i] - 0.01) / (0.1-0.01)) * float32(histresolution-1) + 0.5) ))
        if index < 0 {
          //fmt.Printf("should not happen %d %g \n", index, (field[k][j][i]))
          index = 0
        }
        if index > histresolution-1 {
          //fmt.Printf("should not happen %d\n", index)
          index = histresolution-1
        }
        hist[index] = hist[index] + 1
      }
    }
  }
  
  cumhist := make([]int64, histresolution)
  cumhist[0] = int64(hist[0])
  for i := 1; i < len(cumhist); i++ {
    cumhist[i] = int64(hist[i]) + cumhist[i-1]
    //fmt.Printf("Hist %d %d %d\n", i, int(hist[i]), cumhist[i])
  }
  total := cumhist[len(cumhist)-1]
  // now check quantiles
  thresholds := make([]float32, numsegments-1)
  t := 0
  for i := range cumhist {
     //fmt.Printf("val: %g\n", float64(cumhist[i])/float64(total))
     if float64(cumhist[i])/float64(total) > float64(borders[t]) {
       thresholds[t] = float32(float64(0.01) + float64(0.1-0.01)*(float64(i)/float64(len(cumhist)-1.0)))
       //fmt.Printf("heat threshold value %d is %g hit at index %d\n", t, thresholds[t], i)
       t = t + 1
       if t >= len(borders) {
         break
       }
     }
  }

  // now use the thresholds to compute for each voxel what the region it is in
  for k := range df {
    for j := range df[k] {
      for i := range df[k][j] {
          if simThese[k][j][i] == 0 {
            df[k][j][i] = 0
            continue
          }
          df[k][j][i] = 1
          for l := range thresholds {
            lab := len(thresholds)-1-l
            //fmt.Printf("check threshold: field %g > threshold %g\n", field[k][j][i], thresholds[lab])
            if field[k][j][i] > thresholds[lab] {
              //fmt.Printf("OK, set this to label %d\n", lab+1)
              df[k][j][i] = uint8(lab+1)
              break
            }
          }
          //fmt.Printf("done\n")
      }
    }
  }

  return df  
}

func simulate( labels [][][]uint8, temp0 []int, temp1 []int, simulate []int, omega float32, iterations int, showAllTemps bool, verbose bool) ( [][][]float32 ){
  // write the input field to fn
  var dims [3]int
  dims[2] = len(labels)
  dims[1] = len(labels[0])
  dims[0] = len(labels[0][0])
  f := make([][][]float32, dims[2])
  for i := range f {
    f[i] = make([][]float32, dims[1])
    for j := range f[i] {
      f[i][j] = make([]float32, dims[0])
    }
  }

  // a temporary copy, update at the end of each cycle
  tmp := make([][][]float32, dims[2])
  for i := range tmp {
    tmp[i] = make([][]float32, dims[1])
    for j := range tmp[i] {
      tmp[i][j] = make([]float32, dims[0])
    }
  }

  // memorize what label we do want to simulate (=1), and what labels are repulsive (=2)
  simThese := make([][][]uint8, dims[2])
  for i := range simThese {
    simThese[i] = make([][]uint8, dims[1])
    for j := range simThese[i] {
      simThese[i][j] = make([]uint8, dims[0])
    }
  }
  
  // set the initial temperatures
  for k := 0; k < dims[2]; k++ {
    for j := 0; j < dims[1]; j++ {
      for i := 0; i < dims[0]; i++ {
         f[k][j][i] = 0.0
         simThese[k][j][i] = 2 // don't simulate this label, repulsive boundary conditions
         val := int(labels[k][j][i])
         for l := range temp0 {
           if temp0[l] == val {
             f[k][j][i] = 0.01
             simThese[k][j][i] = 0
           }
         }
         for l := range temp1 {
           if temp1[l] == val {
             f[k][j][i] = 0.1
             simThese[k][j][i] = 0
           }
         }
         for l := range simulate {
           if simulate[l] == val {
             f[k][j][i] = 0.01 + (0.1-0.01)/2.0 // initialize with the mean temperature between 0.01 and 0.1
             simThese[k][j][i] = 1 // yes simulate this voxel
           }
         }         
      }
    }
  }  
  
  // now simulate a couple of iterations
  maxTime := iterations
  for time := 0; time < maxTime; time++ {
    if verbose {
      fmt.Printf("\r%04d/%d", time+1, maxTime)
    }
    for k := 1; k < dims[2]-1; k++ {
       for j := 1; j < dims[1]-1; j++ {
         copy(tmp[k][j][:],f[k][j][:])
         for i := 1; i < dims[0]-1; i++ {
            //tmp[k][j][i] = f[k][j][i]
            if simThese[k][j][i] != 1 {
              continue
            }
            var val111 = f[ k ][j][i]
            var val101 = f[k][j-1][i]
            var val121 = f[k][j+1][i]
            var val011 = f[k][j][i-1]
            var val211 = f[k][j][i+1]
            var val110 = f[k-1][j][i]
            var val112 = f[k+1][j][i]
            // repulsive boundary conditions for all other label
            if labels[k][j-1][i] == 2 {
              val101 = val121
            }
            if labels[k][j+1][i] == 2 {
              val121 = val211
            }
            if labels[k][j][i-1] == 2 {
              val011 = val211
            }
            if labels[k][j][i+1] == 2 {
              val211 = val011
            }
            if labels[k-1][j][i] == 2 {
              val110 = val112
            }
            if labels[k+1][j][i] == 2 {
              val112 = val110
            }
            tmp[k][j][i] = float32(1.0-6.0*omega)*val111 + omega*(val101 + val121 + val011 + val211 + val110 + val112)            
         }
       }
    }   
    // now copy values over to real dataset
    for k := 0; k < dims[2]; k++ {
       for j := 0; j < dims[1]; j++ {
          copy(f[k][j][0:dims[0]], tmp[k][j][:])
       }
    }    
  }
  
  // at the end leave only the simulated voxel in the image
  if ! showAllTemps {
    // remove all entries which are not simulate voxel
    for k := 0; k < dims[2]; k++ {
      for j := 0; j < dims[1]; j++ {
        for i := 0; i < dims[0]; i++ {
            if simThese[k][j][i] != 1 {
              f[k][j][i] = 0
            }
        }
      }
    }
  }
  
  if verbose {
    fmt.Printf("\n")
  }
  return f
}



