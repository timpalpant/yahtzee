# Roadmap

This is a list of all of the functionality areas within OpenCV, and OpenCV Contrib.

Any section listed with an "X" means that all of the relevant OpenCV functionality has been wrapped for use within GoCV.

Any section listed with **WORK STARTED** indicates that some work has been done, but not all functionality in that module has been completed. If there are any functions listed under a section marked **WORK STARTED**, it indicates that that function still requires a wrapper implemented.

And any section that is simply listed, indicates that so far, no work has been done on that module.

Your pull requests will be greatly appreciated!

## Modules list

- [ ] core. Core functionality
    - [ ] **Basic structures - WORK STARTED**
    - [ ] **Operations on arrays - WORK STARTED**. The following functions still need implementation:
        - [ ] [batchDistance](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga4ba778a1c57f83233b1d851c83f5a622)
        - [ ] [borderInterpolate](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga247f571aa6244827d3d798f13892da58)
        - [ ] [calcCovarMatrix](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#gae6ffa9354633f984246945d52823165d)
        - [ ] [cartToPolar](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#gae6ffa9354633f984246945d52823165d)
        - [ ] [checkRange](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga2bd19d89cae59361416736f87e3c7a64)
        - [ ] [compare](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#gae6ffa9354633f984246945d52823165d)
        - [ ] [completeSymm](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#gaa9d88dcd0e54b6d1af38d41f2a3e3d25)
        - [ ] [convertFp16](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga9c25d9ef44a2a48ecc3774b30cb80082)
        - [ ] [convertScaleAbs](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga3460e9c9f37b563ab9dd550c4d8c4e7d)
        - [ ] [copyMakeBorder](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga3460e9c9f37b563ab9dd550c4d8c4e7d)
        - [ ] [countNonZero](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#gaa4b89393263bb4d604e0fe5986723914)
        - [ ] [dct](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga85aad4d668c01fbd64825f589e3696d4)
        - [ ] [determinant](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#gaf802bd9ca3e07b8b6170645ef0611d0c)
        - [ ] [divide](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga6db555d30115642fedae0cda05604874)
        - [ ] [eigen](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga9fa0d58657f60eaa6c71f6fbb40456e3)
        - [ ] [eigenNonSymmetric](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#gaf51987e03cac8d171fbd2b327cf966f6)
        - [ ] [exp](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga3e10108e2162c338f1b848af619f39e5)
        - [ ] [extractChannel](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#gacc6158574aa1f0281878c955bcf35642)
        - [ ] [findNonZero](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#gaed7df59a3539b4cc0fe5c9c8d7586190)
        - [ ] [flip](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#gaca7be533e3dac7feb70fc60635adf441)
        - [ ] [gemm](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#gacb6e64071dffe36434e1e7ee79e7cb35)
        - [ ] [hconcat](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#gacb6e64071dffe36434e1e7ee79e7cb35)
        - [ ] [idct](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga77b168d84e564c50228b69730a227ef2)
        - [ ] [idft](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#gaa708aa2d2e57a508f968eb0f69aa5ff1)
        - [ ] [insertChannel](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga1d4bd886d35b00ec0b764cb4ce6eb515)
        - [ ] [invert](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#gad278044679d4ecf20f7622cc151aaaa2)
        - [ ] [log](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga937ecdce4679a77168730830a955bea7)
        - [ ] [magnitude](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga6d3b097586bca4409873d64a90fe64c3)
        - [ ] [Mahalanobis](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga4493aee129179459cbfc6064f051aa7d)
        - [ ] [max](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#gacc40fa15eac0fb83f8ca70b7cc0b588d)
        - [ ] [meanStdDev](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga846c858f4004d59493d7c6a4354b301d)
        - [ ] [min](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga9af368f182ee76d0463d0d8d5330b764)
        - [ ] [minMaxIdx](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga7622c466c628a75d9ed008b42250a73f)
        - [ ] [mixChannels](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga51d768c270a1cdd3497255017c4504be)
        - [ ] [mulSpectrums](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga3ab38646463c59bf0ce962a9d51db64f)
        - [ ] [multiply](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga979d898a58d7f61c53003e162e7ad89f)
        - [ ] [mulTransposed](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#gadc4e49f8f7a155044e3be1b9e3b270ab)
        - [ ] [patchNaNs](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga62286befb7cde3568ff8c7d14d5079da)
        - [ ] [PCABackProject](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#gab26049f30ee8e94f7d69d82c124faafc)
        - [ ] [PCACompute](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga4e2073c7311f292a0648f04c37b73781)
        - [ ] [PCAProject](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga6b9fbc7b3a99ebfd441bbec0a6bc4f88)
        - [ ] [perspectiveTransform](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#gad327659ac03e5fd6894b90025e6900a7)
        - [ ] [phase](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga9db9ca9b4d81c3bde5677b8f64dc0137)
        - [ ] [polarToCart](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga581ff9d44201de2dd1b40a50db93d665)
        - [ ] [pow](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#gaf0d056b5bd1dc92500d6f6cf6bac41ef)
        - [ ] [PSNR](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga07aaf34ae31d226b1b847d8bcff3698f)
        - [ ] [randn](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#gaeff1f61e972d133a04ce3a5f81cf6808)
        - [ ] [randShuffle](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga6a789c8a5cb56c6dd62506179808f763)
        - [ ] [randu](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga1ba1026dca0807b27057ba6a49d258c0)
        - [ ] [reduce](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga4b78072a303f29d9031d56e5638da78e)
        - [ ] [repeat](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga496c3860f3ac44c40b48811333cfda2d)
        - [ ] [rotate](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga4ad01c0978b0ce64baa246811deeac24)
        - [ ] [scaleAdd](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga9e0845db4135f55dcf20227402f00d98)
        - [ ] [setIdentity](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga388d7575224a4a277ceb98ccaa327c99)
        - [ ] [setRNGSeed](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga757e657c037410d9e19e819569e7de0f)
        - [ ] [solve](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga12b43690dbd31fed96f213eefead2373)
        - [ ] [solveCubic](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga1c3b0b925b085b6e96931ee309e6a1da)
        - [ ] [solvePoly](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#gac2f5e953016fabcdf793d762f4ec5dce)
        - [ ] [sort](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga45dd56da289494ce874be2324856898f)
        - [ ] [sortIdx](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#gadf35157cbf97f3cb85a545380e383506)
        - [ ] [split](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga0547c7fed86152d7e9d0096029c8518a)
        - [ ] [sqrt](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga186222c3919657890f88df5a1f64a7d7)
        - [ ] [subtract](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#gaa0f00d98b4b5edeaeb7b8333b2de353b)
        - [ ] [sum](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga716e10a2dd9e228e4d3c95818f106722)
        - [ ] [SVBackSubst](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#gab4e620e6fc6c8a27bb2be3d50a840c0b)
        - [ ] [SVDecomp](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#gab477b5b7b39b370bb03e75b19d2d5109)
        - [ ] [theRNG](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga75843061d150ad6564b5447e38e57722)
        - [ ] [trace](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga3419ac19c7dcd2be4bd552a23e147dd8)
        - [ ] [transform](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga393164aa54bb9169ce0a8cc44e08ff22)
        - [ ] [transpose](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga46630ed6c0ea6254a35f447289bd7404)
        - [ ] [vconcat](https://docs.opencv.org/3.4.0/d2/de8/group__core__array.html#ga744f53b69f6e4f12156cdde4e76aed27)
    - [ ] XML/YAML Persistence
    - [ ] Clustering
    - [ ] Utility and system functions and macros
    - [ ] OpenGL interoperability
    - [ ] Intel IPP Asynchronous C/C++ Converters
    - [ ] Optimization Algorithms
    - [ ] OpenCL support 

- [ ] imgproc. Image processing
    - [ ] **Image Filtering - WORK STARTED**
    - [ ] **Geometric Image Transformations - WORK STARTED**
    - [ ] **Miscellaneous Image Transformations - WORK STARTED**
    - [ ] **Drawing Functions - WORK STARTED**
    - [ ] ColorMaps in OpenCV
    - [ ] Planar Subdivision
    - [ ] Histograms
    - [ ] Structural Analysis and Shape Descriptors
    - [ ] **Motion Analysis and Object Tracking - WORK STARTED**
    - [ ] **Feature Detection - WORK STARTED**
    - [ ] **Object Detection - WORK STARTED**

- [X] **imgcodecs. Image file reading and writing.**
- [X] **videoio. Video I/O**
- [X] **highgui. High-level GUI**
- [ ] **video. Video Analysis - WORK STARTED**
    - [X] **Motion Analysis**
    - [ ] **Object Tracking - WORK STARTED** (https://docs.opencv.org/3.4.0/dc/d6b/group__video__track.html)

- [ ] calib3d. Camera Calibration and 3D Reconstruction
- [ ] **features2d. 2D Features Framework - WORK STARTED**
    - [X] **Feature Detection and Description**
    - [ ] Descriptor Matchers (https://docs.opencv.org/3.4.0/d8/d9b/group__features2d__match.html)
    - [ ] Drawing Function of Keypoints and Matches (https://docs.opencv.org/3.4.0/d4/d5d/group__features2d__draw.html)
    - [ ] Object Categorization (https://docs.opencv.org/3.4.0/de/d24/group__features2d__category.html)

- [X] **objdetect. Object Detection**
- [ ] **dnn. Deep Neural Network module - WORK STARTED**
- [ ] ml. Machine Learning
- [ ] flann. Clustering and Search in Multi-Dimensional Spaces
- [ ] photo. Computational Photography
- [ ] stitching. Images stitching
- [ ] cudaarithm. Operations on Matrices
- [ ] cudabgsegm. Background Segmentation
- [ ] cudacodec. Video Encoding/Decoding
- [ ] cudafeatures2d. Feature Detection and Description
- [ ] cudafilters. Image Filtering
- [ ] cudaimgproc. Image Processing
- [ ] cudalegacy. Legacy support
- [ ] cudaobjdetect. Object Detection
- [ ] cudaoptflow. Optical Flow
- [ ] cudastereo. Stereo Correspondence
- [ ] cudawarping. Image Warping
- [ ] cudev. Device layer
- [ ] shape. Shape Distance and Matching
- [ ] superres. Super Resolution
- [ ] videostab. Video Stabilization
- [ ] viz. 3D Visualizer

## Contrib modules list

- [ ] aruco. ArUco Marker Detection
- [ ] bgsegm. Improved Background-Foreground Segmentation Methods
- [ ] bioinspired. Biologically inspired vision models and derivated tools
- [ ] ccalib. Custom Calibration Pattern for 3D reconstruction
- [ ] cnn_3dobj. 3D object recognition and pose estimation API
- [ ] cvv. GUI for Interactive Visual Debugging of Computer Vision Programs
- [ ] datasets. Framework for working with different datasets
- [ ] dnn_modern. Deep Learning Modern Module
- [ ] dpm. Deformable Part-based Models
- [ ] **face. Face Recognition - WORK STARTED**
- [ ] freetype. Drawing UTF-8 strings with freetype/harfbuzz
- [ ] fuzzy. Image processing based on fuzzy mathematics
- [ ] hdf. Hierarchical Data Format I/O routines
- [ ] img_hash. The module brings implementations of different image hashing algorithms.
- [ ] line_descriptor. Binary descriptors for lines extracted from an image
- [ ] matlab. MATLAB Bridge
- [ ] optflow. Optical Flow Algorithms
- [ ] phase_unwrapping. Phase Unwrapping API
- [ ] plot. Plot function for Mat data
- [ ] reg. Image Registration
- [ ] rgbd. RGB-Depth Processing
- [ ] saliency. Saliency API
- [ ] sfm. Structure From Motion
- [ ] stereo. Stereo Correspondance Algorithms
- [ ] structured_light. Structured Light API
- [ ] surface_matching. Surface Matching
- [ ] text. Scene Text Detection and Recognition
- [ ] tracking. Tracking API
- [ ] **xfeatures2d. Extra 2D Features Framework - WORK STARTED**
- [ ] ximgproc. Extended Image Processing
- [ ] xobjdetect. Extended object detection
- [ ] xphoto. Additional photo processing algorithms
