#include "core.h"
#include <string.h>

// Mat_New creates a new empty Mat
Mat Mat_New() {
    return new cv::Mat();
}

// Mat_NewWithSize creates a new Mat with a specific size dimension and number of channels.
Mat Mat_NewWithSize(int rows, int cols, int type) {
    return new cv::Mat(rows, cols, type, 0.0);
}

// Mat_NewFromScalar creates a new Mat from a Scalar. Intended to be used
// for Mat comparison operation such as InRange.
Mat Mat_NewFromScalar(Scalar ar, int type) {
    cv::Scalar c = cv::Scalar(ar.val1, ar.val2, ar.val3, ar.val4);
    return new cv::Mat(1, 1, type, c);
}

Mat Mat_NewFromBytes(int rows, int cols, int type, struct ByteArray buf) {
    return new cv::Mat(rows, cols, type, buf.data);
}

// Mat_Close deletes an existing Mat
void Mat_Close(Mat m) {
    delete m;
}

// Mat_Empty tests if a Mat is empty
int Mat_Empty(Mat m) {
    return m->empty();
}

// Mat_Clone returns a clone of this Mat
Mat Mat_Clone(Mat m) {
    return new cv::Mat(m->clone());
}

// Mat_CopyTo copies this Mat to another Mat.
void Mat_CopyTo(Mat m, Mat dst) {
    m->copyTo(*dst);
}

void Mat_ConvertTo(Mat m, Mat dst, int type) {
    m->convertTo(*dst, type);
}

// Mat_ToBytes returns the bytes representation of the underlying data.
struct ByteArray Mat_ToBytes(Mat m) {
        return toByteArray(reinterpret_cast<const char*>(m->data), m->total() * m->elemSize());
}

// Mat_Region returns a Mat of a region of another Mat
Mat Mat_Region(Mat m, Rect r) {
    return new cv::Mat(*m, cv::Rect(r.x, r.y, r.width, r.height));
}

Mat Mat_Reshape(Mat m, int cn, int rows) {
    return new cv::Mat(m->reshape(cn, rows));
}

Mat Mat_ConvertFp16(Mat m) {
    Mat dst = new cv::Mat();
    cv::convertFp16(*m, *dst);
    return dst;
}

// Mat_Mean calculates the mean value M of array elements, independently for each channel, and return it as Scalar vector
// TODO pass second paramter with mask
Scalar Mat_Mean(Mat m) {
    cv::Scalar c = cv::mean(*m);
    Scalar scal = Scalar();
    scal.val1 = c.val[0];
    scal.val2 = c.val[1];
    scal.val3 = c.val[2];
    scal.val4 = c.val[3];
    return scal;
}

void LUT(Mat src, Mat lut, Mat dst) {
  cv::LUT(*src, *lut, *dst);
}

// Mat_Rows returns how many rows in this Mat.
int Mat_Rows(Mat m) {
    return m->rows;
}

// Mat_Cols returns how many columns in this Mat.
int Mat_Cols(Mat m) {
    return m->cols;
}

// Mat_Channels returns how many channels in this Mat.
int Mat_Channels(Mat m) {
    return m->channels();
}

// Mat_Type returns the type from this Mat.
int Mat_Type(Mat m) {
    return m->type();
}

// Mat_GetUChar returns a specific row/col value from this Mat expecting
// each element to contain a schar aka CV_8U.
uint8_t Mat_GetUChar(Mat m, int row, int col) {
    return m->at<uchar>(row, col);
}

// Mat_GetSChar returns a specific row/col value from this Mat expecting
// each element to contain a schar aka CV_8S.
int8_t Mat_GetSChar(Mat m, int row, int col) {
    return m->at<schar>(row, col);
}

// Mat_GetShort returns a specific row/col value from this Mat expecting
// each element to contain a short aka CV_16S.
int16_t Mat_GetShort(Mat m, int row, int col) {
    return m->at<short>(row, col);
}

// Mat_GetInt returns a specific row/col value from this Mat expecting
// each element to contain an int aka CV_32S.
int32_t Mat_GetInt(Mat m, int row, int col) {
    return m->at<int>(row, col);
}

// Mat_GetFloat returns a specific row/col value from this Mat expecting
// each element to contain a float aka CV_32F.
float Mat_GetFloat(Mat m, int row, int col) {
    return m->at<float>(row, col);
}

// Mat_GetDouble returns a specific row/col value from this Mat expecting
// each element to contain a double aka CV_64F.
double Mat_GetDouble(Mat m, int row, int col) {
    return m->at<double>(row, col);
}

// Mat_SetUChar set a specific row/col value from this Mat expecting
// each element to contain a schar aka CV_8U.
void Mat_SetUChar(Mat m, int row, int col, uint8_t val) {
    m->at<uchar>(row, col) = val;
}

// Mat_SetSChar set a specific row/col value from this Mat expecting
// each element to contain a schar aka CV_8S.
void Mat_SetSChar(Mat m, int row, int col, int8_t val) {
  m->at<schar>(row, col) = val;
}

// Mat_SetShort set a specific row/col value from this Mat expecting
// each element to contain a short aka CV_16S.
void Mat_SetShort(Mat m, int row, int col, int16_t val) {
    m->at<short>(row, col) = val;
}

// Mat_SetInt set a specific row/col value from this Mat expecting
// each element to contain an int aka CV_32S.
void Mat_SetInt(Mat m, int row, int col, int32_t val) {
    m->at<int>(row, col) = val;
}

// Mat_SetFloat set a specific row/col value from this Mat expecting
// each element to contain a float aka CV_32F.
void Mat_SetFloat(Mat m, int row, int col, float val) {
    m->at<float>(row, col) = val;
}

// Mat_SetDouble set a specific row/col value from this Mat expecting
// each element to contain a double aka CV_64F.
void Mat_SetDouble(Mat m, int row, int col, double val) {
    m->at<double>(row, col) = val;
}

void Mat_AbsDiff(Mat src1, Mat src2, Mat dst) {
    cv::absdiff(*src1, *src2, *dst);
}

void Mat_Add(Mat src1, Mat src2, Mat dst) {
    cv::add(*src1, *src2, *dst);
}

void Mat_AddWeighted(Mat src1, double alpha, Mat src2, double beta, double gamma, Mat dst) {
    cv::addWeighted(*src1, alpha, *src2, beta, gamma, *dst);
}

void Mat_BitwiseAnd(Mat src1, Mat src2, Mat dst) {
    cv::bitwise_and(*src1, *src2, *dst);
}

void Mat_BitwiseNot(Mat src1, Mat dst) {
    cv::bitwise_not(*src1, *dst);
}

void Mat_BitwiseOr(Mat src1, Mat src2, Mat dst) {
    cv::bitwise_or(*src1, *src2, *dst);
}

void Mat_BitwiseXor(Mat src1, Mat src2, Mat dst) {
    cv::bitwise_xor(*src1, *src2, *dst);
}

void Mat_InRange(Mat src, Mat lowerb, Mat upperb, Mat dst) {
    cv::inRange(*src, *lowerb, *upperb, *dst);
}

int Mat_GetOptimalDFTSize(int vecsize) {
    return cv::getOptimalDFTSize(vecsize);
}

void Mat_DFT(Mat m, Mat dst) {
    cv::dft(*m, *dst);
}

void Mat_Merge(struct Mats mats, Mat dst) {
    std::vector<cv::Mat> images;
    for (int i = 0; i < mats.length; ++i) {
        images.push_back(*mats.mats[i]);
    }
    cv::merge(images, *dst);
}

void Mat_MinMaxLoc(Mat m, double* minVal, double* maxVal, Point* minLoc, Point* maxLoc) {
    cv::Point cMinLoc;
    cv::Point cMaxLoc;
    cv::minMaxLoc(*m, minVal, maxVal, &cMinLoc, &cMaxLoc);

    minLoc->x = cMinLoc.x;
    minLoc->y = cMinLoc.y;
    maxLoc->x = cMaxLoc.x;
    maxLoc->y = cMaxLoc.y;
}

void Mat_Normalize(Mat src, Mat dst, double alpha, double beta, int typ) {
    cv::normalize(*src, *dst, alpha, beta, typ);
}

double Norm(Mat src1, int normType) {
    return cv::norm(*src1, normType);
}

// TermCriteria_New creates a new TermCriteria
TermCriteria TermCriteria_New(int typ, int maxCount, double epsilon) {
    return new cv::TermCriteria(typ, maxCount, epsilon);
}

void Contours_Close(struct Contours cs) {
    for (int i = 0; i < cs.length; i++) {
        Points_Close(cs.contours[i]);
    }
    delete[] cs.contours;
}

void KeyPoints_Close(struct KeyPoints ks) {
    delete[] ks.keypoints;
}

void Points_Close(Points ps) {
    for (size_t i = 0; i < ps.length; i++) {
        Point_Close(ps.points[i]);
    }
    delete[] ps.points;
}

void Point_Close(Point p) {}

void Rects_Close(struct Rects rs) {
    delete[] rs.rects;
}

void Mats_Close(struct Mats mats) {
    delete[] mats.mats;
}

void ByteArray_Release(struct ByteArray buf) {
    delete[] buf.data;
}

struct ByteArray toByteArray(const char* buf, int len) {
    ByteArray ret = {new char[len], len};
    memcpy(ret.data, buf, len);
    return ret;
}
