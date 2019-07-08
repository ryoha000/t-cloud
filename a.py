# coding: UTF-8
import csv
from urllib.request import urlopen
from bs4 import BeautifulSoup
import ssl
ssl._create_default_https_context = ssl._create_unverified_context

# URLの指定
html = urlopen("https://a.sofmap.com/search_result.aspx?mode=SEARCH&gid=&gid=&gid=&keyword_and=&keyword_or=&keyword_not=&product_maker=&product_name=&product_code=&jan_code=4525189004015&product_type=NEW&product_type=USED&price_from=&price_to=&sale_date_from_year=&sale_date_from_month=&sale_date_from_day=&sale_date_to_year=&sale_date_to_month=&sale_date_to_day=&reserve_date_from_year=&reserve_date_from_month=&reserve_date_from_day=&reserve_date_to_year=&reserve_date_to_month=&reserve_date_to_day=&order_by=DEFAULT&styp=p_kwsk")
bsObj = BeautifulSoup(html, "html.parser")
# CSSセレクターを使って指定した場所のtextを表示します
print (bsObj.select_one("#contents_main > form > table:nth-of-type(3) > tr > td > div.search-img-textbox > div.normal-price"))