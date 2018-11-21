# -*- coding: utf-8 -*-

import sys
import json
import urllib2
from urllib import quote

from bs4 import BeautifulSoup

TYPE_MAPPING = {
    "book": {
        "url": "https://m.douban.com/search/?type=book&query={}",
        "cat": "1001"
    },
    "movie": {
        "url": "https://m.douban.com/search/?type=movie&query={}",
        "cat": "1002"
    }
}


def get_items(query, search_type):
    url = TYPE_MAPPING.get(search_type).get("url")
    r = urllib2.urlopen(url.format(quote(query)))
    html = r.read()
    bs = BeautifulSoup(html)
    return bs.select("ul.search_results_subjects > li")


def generate_response(items, search_type):
    base_url = u"https://{}.douban.com".format(search_type)
    cat = TYPE_MAPPING.get(search_type).get("cat")

    result = []
    for item in items:
        href = item.a["href"].replace("/" + search_type, "")
        url = u"{}{}".format(base_url, href)
        origin_score = item.a.div.p.find_all("span")[-1].text
        try:
            score = int(round(float(origin_score)))
        except:
            score = 0

        try:
            midd = u"⚡" if 1.5 > float(origin_score) % 2 >= 0.5 else ""
        except:
            midd = u""

        result.append({
            "type": "file",
            "title": item.a.div.span.text,
            "subtitle": u"⭐" * (score/2) + midd + " " + origin_score,
            "arg": url,
            "icon": {
                "path": "imgs/{}.png".format(search_type)
            }
        })

    result.append({
        "type": "file",
        "title": "more",
        "arg": "{}/subject_search?search_text={}&cat={}".format(base_url, query, cat),
        "icon": {
            "path": "imgs/more.png"
        }
    }
    )

    sys.stdout.write(json.dumps({"items": result}))


if __name__ == "__main__":
    search_type = sys.argv[1]
    query = " ".join(sys.argv[2:])
    items = get_items(query, search_type)
    generate_response(items, search_type)
