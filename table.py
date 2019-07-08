# coding:utf-8
from selenium import webdriver
from selenium.webdriver.chrome.options import Options
import time
import csv
import ssl
ssl._create_default_https_context = ssl._create_unverified_context
from bs4 import BeautifulSoup
from urllib import request

options = Options()
# options.binary_location = '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'

driver = webdriver.Chrome(chrome_options=options)
driver.get('https://erogamescape.dyndns.org/~ap2/ero/toukei_kaiseki/sql_for_erogamer_form.php')

# ID/PASSを入力
sql = driver.find_element_by_name("sql")
sql.clear()
sql.send_keys("SELECT * FROM gamelist WHERE id <'10'")



# ログインボタンをクリック
login_button = driver.find_element_by_css_selector("#submit > input[type=submit]")
login_button.click()


html = driver.page_source.encode('utf-8')
soup = BeautifulSoup(html, "html.parser")

table = soup.select_one("#query_result_main > table")
rows = table.findAll("tr")

with open("ebooks.csv", "w", encoding='utf-8') as file:
    writer = csv.writer(file)
    for row in rows:
        csvRow = []
        for cell in row.findAll(['td', 'th']):
            csvRow.append(cell.get_text())
        writer.writerow(csvRow)

