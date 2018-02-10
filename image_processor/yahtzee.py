import logging
import numpy as np

import skimage.color
import skimage.exposure
import skimage.feature
import skimage.io
import skimage.morphology
import skimage.transform


logger = logging.getLogger(__name__)


def extract_dice(image):
    bw = preprocess(image)
    regions = extract_dice_regions(bw)
    logger.debug("Selected %d regions from image", len(regions))
    result = []
    for region in regions:
        minr, minc, maxr, maxc = region.bbox
        die_image = skimage.exposure.rescale_intensity(bw[minr:maxr, minc:maxc])
        logger.debug("Classifying region %d:%d x %d:%d", minr, maxr, minc, maxc)
        die = classify_die(die_image)
        result.append(die)
    return result


def classify_die(die_image):
    max = None
    max_corr = 0.0

    for candidate, template in TEMPLATES.items():
        x = skimage.transform.resize(die_image, template.shape)
        r = np.corrcoef(x.ravel(), template.ravel())[0, 1]
        logger.debug("Checking %s - R = %.2f", candidate, r)
        if r > max_corr:
            max = candidate
            max_corr = r
    return max


def preprocess(image):
    image = skimage.exposure.equalize_adapthist(image)
    bw = skimage.color.rgb2gray(image)
    return bw


def extract_dice_regions(image):
    edges = skimage.feature.canny(image, sigma=2)
    cleared = skimage.segmentation.clear_border(edges)
    label_image = skimage.measure.label(cleared)
    regions = skimage.measure.regionprops(label_image)
    logger.debug("Found %d regions in labeled image", len(regions))
    return select_dice_regions(regions)


def select_dice_regions(regions: list):
    # Select regions that are approximately square, in a line.
    # NOTE: The regions need to be returned in the correct order,
    # left-to-right.
    result = []
    for r in regions:
        if r.area < 150 or r.area > 250:
            continue
        aspect_ratio = r.major_axis_length / r.minor_axis_length
        if aspect_ratio > 1.5:
            continue
        result.append(r)
    return result


def load_template(filename: str):
    image = skimage.io.imread(filename, as_grey=True)
    w, x, y, z = CROP_REGION
    return image[w:x, y:z]


CROP_REGION = (81, 430, 81, 430)
TEMPLATES = {
    die: load_template('templates/{}.png'.format(die))
    for die in range(1, 7)
}
