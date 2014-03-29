from newshound import fetch
from newshound import stats

import sys
import getopt

help_message = '''
Please run with one of the following options:
    -f,--fetch               'fetch mail from inbox, parse it, save it and match it'
    -r,--rebuild             'reparse alerts and rebuild events'
    -m,--mapreduce           'kicks off mapreduce job for report statistics'
'''
class Usage(Exception):
    def __init__(self, msg):
        self.msg = msg

def raiseUsage():
    raise Usage(help_message)

def main(argv=None):
    if argv is None:
        argv = sys.argv
    try:
        try:
            opts, args = getopt.getopt(argv[1:], "hfrm", ["help", "fetch", "rebuild", "mapreduce"])
        except getopt.error, msg:
            raise Usage(msg)

        # option processing
        if len(opts) == 0:
            raiseUsage()

        for option, value in opts:
            if option in ("-h", "--help"):
                raiseUsage()

            fetcher = fetch.EmailFetcher()

            if option in ("-f", "--fetch"):
                fetcher.get_the_mail()
            elif option in ("-r", "--rebuild"):
                fetcher.rebuild_events()
            elif option in ("-m", "--mapreduce"):
                stats.run()

    except Usage, err:
        print >> sys.stderr, sys.argv[0].split("/")[-1] + ": " + str(err.msg)
        print >> sys.stderr, "for help use --help"
        return 2


if __name__ == '__main__':
    main()
