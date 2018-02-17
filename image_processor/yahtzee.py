import itertools
import logging
import numpy as np

import skimage.color
import skimage.exposure
import skimage.feature
import skimage.io
import skimage.transform


logger = logging.getLogger(__name__)


def extract_dice(image):
    viewport = extract_viewport(image)
    bw = preprocess(viewport)
    regions = extract_dice_regions(bw)
    logger.debug("Selected %d regions from image", len(regions))

    # NOTE: The regions need to be returned in the correct order,
    # left-to-right.
    result = []
    for region in sorted(regions, key=lambda r: r.bbox[1]):
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


def extract_viewport(image):
    bw = preprocess(image)
    edges = skimage.feature.canny(bw, sigma=3)
    cleared = skimage.segmentation.clear_border(edges)
    label_image = skimage.measure.label(cleared)
    regions = skimage.measure.regionprops(label_image)
    largest_region = max(regions, key=lambda r: r.area)
    minr, minc, maxr, maxc = largest_region.bbox
    return image[minr:maxr, minc:maxc]


def extract_dice_regions(image):
    edges = skimage.feature.canny(image, sigma=2)
    cleared = skimage.segmentation.clear_border(edges)
    label_image = skimage.measure.label(cleared)
    regions = skimage.measure.regionprops(label_image)
    logger.debug("Found %d regions in labeled image", len(regions))
    return select_dice_regions(regions)


def select_dice_regions(regions: list):
    # Select regions that are approximately square, in a line.
    candidates = []
    for r in regions:
        minr, minc, maxr, maxc = r.bbox
        height = maxr - minr
        width = maxc - minc
        aspect_ratio = width / height
        if 0.9 < aspect_ratio < 1.1 and 200 < r.area < 500:
            candidates.append(r)
    candidates = remove_overlapping(candidates)
    return squares_in_a_line(candidates, 5)


def remove_overlapping(candidates: list):
    remaining = []
    for region1 in sorted(candidates, key=area):
        area1 = area(region1)
        for region2 in remaining:
            area2 = area(region2)
            if overlap_area(region1, region2) > area1 / 2.0:
                break
        else:
            remaining.append(region1)
    return remaining


def area(region):
    minr, minc, maxr, maxc = region.bbox
    height = maxr - minr
    width = maxc - minc
    return width * height


def overlap_area(region1, region2):
    minr1, minc1, maxr1, maxc1 = region1.bbox
    minr2, minc2, maxr2, maxc2 = region2.bbox
    minr = max(minr1, minr2)
    minc = max(minc1, minc2)
    maxr = min(maxr1, maxr2)
    maxc = min(maxc1, maxc2)
    height = max(maxr - minr, 0)
    width = max(maxc - minc, 0)
    return width * height


def squares_in_a_line(candidates: list, n: int):
    # Of the given candidates, return the n that are most linear.
    if len(candidates) < n:
        return candidates

    scores = []
    for regions in itertools.combinations(candidates, n):
        regions = sorted(regions, key=lambda r: r.bbox)
        widths = [r.bbox[3] - r.bbox[1] for r in regions]
        heights = [r.bbox[2] - r.bbox[0] for r in regions]
        horizontal_spacing = [regions[i+1].bbox[1] - regions[i].bbox[1]
                              for i in range(len(regions)-1)]
        vertical_spacing = [regions[i+1].bbox[2] - regions[i].bbox[0]
                            for i in range(len(regions)-1)]
        sizing_var = np.var(widths) + np.var(heights)
        spacing_var = np.var(horizontal_spacing) + np.var(vertical_spacing)
        total_var = sizing_var + spacing_var
        scores.append((total_var, regions))

    return min(scores)[1]


def load_template(filename: str):
    image = skimage.io.imread(filename, as_grey=True)
    w, x, y, z = CROP_REGION
    return image[w:x, y:z]


CROP_REGION = (81, 430, 81, 430)
TEMPLATES = {
    die: load_template('templates/{}.png'.format(die))
    for die in range(1, 7)
}
