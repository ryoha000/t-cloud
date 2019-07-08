#coding: UTF-8
from urllib import request
from bs4 import BeautifulSoup
from selenium import webdriver
from selenium.webdriver.chrome.options import Options
import time
import csv


# def scraping():
for i in range(16301,16303):
    #url
    options = Options()
    # options.set_headless(True)
    driver = webdriver.Chrome(chrome_options=options)
    # i=17546

    driver.get("https://erogamescape.dyndns.org/~ap2/ero/toukei_kaiseki/mod.php?game=%d#change" % i)

    html = driver.page_source.encode('utf-8')

    

    #set BueatifulSoup
    soup = BeautifulSoup(html, "html.parser")

    #get headlines
    # mainNewsIndex = soup.find_all("span", attrs={"class", "clip_bd"})
    time.sleep(0.8)

    asin = ['']*4
    jan = ['']*4
    if soup.select_one("body > div.scape > form:nth-child(14) > p:nth-child(12) > a:nth-child(2)") == None:
        break
    else:

        asin[0] = soup.select_one("body > div.scape > form:nth-child(14) > p:nth-child(12) > a:nth-child(2)").string
        asin[1] = soup.select_one("body > div.scape > form:nth-child(14) > p:nth-child(12) > a:nth-child(5)").string
        asin[2] = soup.select_one("body > div.scape > form:nth-child(14) > p:nth-child(12) > a:nth-child(8)").string
        asin[3] = soup.select_one("body > div.scape > form:nth-child(14) > p:nth-child(12) > a:nth-child(11)").string

        for v in range(0,4):
             if asin[v] == None:
                 break
             else:
                 driver.get("https://mnrate.com/item/aid/%s" % asin[v])
                 html = driver.page_source.encode('utf-8')
                 soup = BeautifulSoup(html, "html.parser")
                 jan[v] = soup.select_one("#main_contents > div.list_style_default_parent.displayRateComponent > ul.item_box > li.item_detail > div:nth-child(3) > span:nth-child(2) > span.clip_bd").string
    
        with open("id-asin-jan.csv", "a", encoding='utf-8',newline='') as f:
            writer = csv.writer(f, lineterminator='\n')
            writer.writerow([i, asin[0], asin[1], asin[2], asin[3], jan[0], jan[1], jan[2], jan[3]])
        

else:
    print('finish')

# if __name__ == "__main__":
#     scraping()

    # <span id="_asin" class="clip_bd" data-clipboard-text="B00USS9QJK" style="cursor: pointer;">B00USS9QJK</span>