#!/usr/bin/env python
import base64
import flask
import io
import logging
import skimage.io

import yahtzee


app = flask.Flask(__name__)

logger = logging.getLogger()
gunicorn_error_logger = logging.getLogger('gunicorn.error')
logger.handlers.extend(gunicorn_error_logger.handlers)
logger.setLevel(logging.DEBUG)


@app.route("/rest/v1/process_image", methods=["POST"])
def process_image():
    '''
    Process an image of electronic hand-held Yahtzee,
    extracting the current dice roll.

    Request: {
        "image": "base64-encoded image",
    }

    Response: {
        "dice": [1, 2, 3, 4, 5],
        "error": {
            "message": "Could not extract dice",
        },
    }
    '''
    try:
        req = flask.request.get_json(force=True)
        buf = base64.b64decode(req["image"], validate=True)
        a = skimage.io.imread(io.BytesIO(buf))
    except Exception as e:
        resp = flask.make_response(str(e), 400)
        flask.abort(resp)

    logging.info("Extracting dice from image")
    resp = {}
    try:
        resp["dice"] = yahtzee.extract_dice(a)
    except Exception as e:
        resp["error"] = {"message": str(e)}

    return flask.jsonify(resp)


if __name__ == '__main__':
    app.debug = True
    app.run(host='localhost', port=8080)
