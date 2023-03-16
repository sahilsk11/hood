import lambda_email_parser
from sample_events import *

def test_tda_event_parser():
    lambda_email_parser.lambda_handler(tda_event(), None)

test_tda_event_parser()