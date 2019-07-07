#coding: UTF-8
from urllib import request
from bs4 import BeautifulSoup
from selenium import webdriver
from selenium.webdriver.chrome.options import Options
import time

def scraping():
    options = Options()
    # options.set_headless(True)
    driver = webdriver.Chrome(chrome_options=options)
    #url
    driver.get("https://erogamescape.dyndns.org/~ap2/ero/toukei_kaiseki/game.php?game=17546")

    html = driver.page_source.encode('utf-8')

    

    #set BueatifulSoup
    soup = BeautifulSoup(html, "html.parser")

    #get headlines
    # mainNewsIndex = soup.find_all("span", attrs={"class", "clip_bd"})
    time.sleep(1)
    a = soup.select_one("#median > td")
    time.sleep(0.5)

    #print headlines
    print(a)

if __name__ == "__main__":
    scraping()

    # <span id="_asin" class="clip_bd" data-clipboard-text="B00USS9QJK" style="cursor: pointer;">B00USS9QJK</span>