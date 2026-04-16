import os
import json
import time
import requests
from dotenv import load_dotenv
from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC

import tkinter as tk
from tkinter import messagebox

import tkinter as tk
from tkinter import messagebox


def validate_years(start: str, end: str):
	if not start.isdigit() or not end.isdigit():
		return False, "Please enter numbers only."

	start = int(start)
	end = int(end)

	if start == None or end == None:
		return False, "Please enter a date!"

	if start < 2009:
		return False, "Start year must be 2009 or greater."

	if start > end:
		return False, "Start year must be less than or equal to end year."

	return True, (start, end)


def enem_year_range_window():
	result = {"start": None, "end": None}

	def submit():
		valid, data = validate_years(entry_start.get(), entry_end.get())

		if not valid:
			messagebox.showerror("Error", data)
			return

		result["start"], result["end"] = data
		root.destroy()

	def on_close():
		result["start"] = None
		result["end"] = None
		root.destroy()

	root = tk.Tk()
	root.title("ENEM - Year Range")
	root.geometry("300x180")
	root.resizable(False, False)
	root.protocol("WM_DELETE_WINDOW", on_close)
	tk.Label(root, text="Select ENEM year range", font=("Arial", 11, "bold")).pack(pady=10)

	frame = tk.Frame(root)
	frame.pack()

	tk.Label(frame, text="Start year:").grid(row=0, column=0, padx=5, pady=5)
	entry_start = tk.Entry(frame, width=10)
	entry_start.grid(row=0, column=1)

	tk.Label(frame, text="End year:").grid(row=1, column=0, padx=5, pady=5)
	entry_end = tk.Entry(frame, width=10)
	entry_end.grid(row=1, column=1)

	tk.Button(root, text="Confirm", command=submit).pack(pady=15)

	root.wait_window()

	return result["start"], result["end"]


load_dotenv()

def config_webdriver() -> webdriver.Chrome:
	options = Options()
	options.add_argument("--start-maximized")
	options.add_argument("--disable-gpu")
	options.add_argument("--no-sandbox")
	options.add_argument("--disable-dev-shm-usage")
	driver = webdriver.Chrome(options=options)
	driver.set_page_load_timeout(120)
	return driver

def download_img(url: str, folder_path: str, file_name: str) -> str:
	if not url or not url.startswith('http'):
		return None
	try:
		os.makedirs(folder_path, exist_ok=True)
		file_extension = url.split('.')[-1].split('?')[0].lower()
		if file_extension not in ['png', 'jpg', 'jpeg', 'gif', 'webp']:
			file_extension = "png"

		full_name = f"{file_name}.{file_extension}"
		full_path = os.path.join(folder_path, full_name)

		headers = {'User-Agent': 'Mozilla/5.0'}
		response = requests.get(url, headers=headers, stream=True, timeout=10)

		if response.status_code == 200:
			with open(full_path, 'wb') as f:
				for chunk in response.iter_content(1024):
					f.write(chunk)
			return full_path.replace("\\", "/")
	except Exception as e:
		print(f"Error img {file_name}: {e}")
	return None

def main():
	email = os.getenv("ENEM_EMAIL")
	password = os.getenv("ENEM_PASSWORD")
	start_year, end_year = enem_year_range_window()
	if start_year is None or end_year is None:
		print("Program cancelled by user.")
		return
	driver = config_webdriver()
	wait = WebDriverWait(driver, 20)

	base_path = "banco_de_questoes"
	organized_data = {}
	exam_types = ["ENEM PPL", "ENEM DIGITAL", "ENEM - Pará", "ENEM (Libras)", "ENEM 2° EDIÇÃO", "ENEM"]

	try:
		driver.get("https://app.repertorioenem.com.br")
		wait.until(EC.presence_of_element_located((By.ID, "inputEmailAddress"))).send_keys(email)
		driver.find_element(By.ID, "inputPassword").send_keys(password)
		driver.find_element(By.CSS_SELECTOR, "button[type='submit']").click()
		time.sleep(5)

		for year in range(start_year, end_year + 1):
			page = 1
			year_str = str(year)
			while True:
				print(f"Processando {year_str} - Página {page}...")
				url = f"https://app.repertorioenem.com.br/questions/list?search=1&institution%5B%5D=1&institution%5B%5D=204&institution%5B%5D=168&institution%5B%5D=157&institution%5B%5D=3&institution%5B%5D=2&year%5B%5D={year}&pages=100&page={page}"
				driver.get(url)

				try:
					wait.until(EC.presence_of_element_located((By.CLASS_NAME, "card.my-1")))
				except:
					break

				js_extractor = """
				let items = [];

				const limparEspacosExtra = (txt) => {
					if (!txt) return "";
					return txt.replace(/[ ]{2,}/g, ' ').trim();
				};

				document.querySelectorAll('.card.my-1').forEach(card => {
					try {
						let idFull = card.querySelector('.h6.mb-2.mx-2').innerText.trim();
						let idNum = idFull.replace('Q', '');
						let divEnunciado = card.querySelector('.ck-content');
						let enunciadoBruto = "";
						let imgEnunciadoUrls = [];
						let contadorImg = 1;

						const processNode = (node, isGabarito = false) => {
							let textBuffer = "";
							if (node.nodeType === Node.TEXT_NODE) {
								textBuffer += node.textContent;
							} else if (node.nodeType === Node.ELEMENT_NODE) {
								if (isGabarito && (node.classList.contains('VideoAnswerModal') || node.tagName === 'BUTTON')) {
									return "";
								}

								if (node.tagName === 'IMG') {
									if (!isGabarito) {
										textBuffer += `\\n[[REF_IMG_${contadorImg}]]\\n`;
										imgEnunciadoUrls.push(node.src);
										contadorImg++;
									}
								} else if (node.tagName === 'BR') {
									textBuffer += "\\n";
								} else {
									node.childNodes.forEach(child => {
										textBuffer += processNode(child, isGabarito);
									});
									if (['P', 'DIV', 'BLOCKQUOTE', 'LI', 'H4', 'H5'].includes(node.tagName)) {
										textBuffer += "\\n";
									}
								}
							}
							return textBuffer;
						};

						if (divEnunciado) {
							enunciadoBruto = processNode(divEnunciado, false);
						}

						let divGabarito = card.querySelector('.card.card-body.mt-2');
						let gabComentadoBruto = "";
						if (divGabarito) {
							gabComentadoBruto = processNode(divGabarito, true);
							gabComentadoBruto = gabComentadoBruto.split("Vídeo de Resolução da Questão")[0];
						}

						let todasTags = Array.from(card.querySelectorAll('.badge')).map(t => t.innerText.trim());
						let dificuldade = "Não Definida";
						const niveis = ["Fácil", "Média", "Médio", "Difícil", "Muito Fácil", "Muito Difícil"];
						let tagsFiltradas = todasTags.filter(tag => {
							if (niveis.some(n => tag.toLowerCase() === n.toLowerCase())) {
								dificuldade = tag; return false;
							}
							return true;
						});

						let altTexto = {};
						let altImgUrls = {};
						['a', 'b', 'c', 'd', 'e'].forEach(letra => {
							let label = document.getElementById(`check_${letra}_${idNum}`);
							if (label) {
								let txtRaw = label.innerText.replace(/^[a-eA-E]\\s?\\)\\s?/, "");
								altTexto[letra.toUpperCase()] = limparEspacosExtra(txtRaw);
								let img = label.querySelector('img');
								if (img) altImgUrls[letra.toUpperCase()] = img.src;
							}
						});

						items.push({
							id: idFull,
							id_numero: idNum,
							dificuldade: dificuldade,
							enunciado: limparEspacosExtra(enunciadoBruto),
							imgs_enunciado_urls: imgEnunciadoUrls,
							alternativas_texto: altTexto,
							alternativas_imgs_urls: altImgUrls,
							gabarito: document.getElementById('correctAlternative' + idNum)?.value.toUpperCase() || "N/D",
							gabarito_comentado: limparEspacosExtra(gabComentadoBruto),
							tags: tagsFiltradas
						});
					} catch(e) {}
				});
				return items;
				"""

				questions_raw = driver.execute_script(js_extractor)
				if not questions_raw:
					break

				for q in questions_raw:
					folder_q = os.path.join(base_path, year_str, q['id'])

					local_imgs_subject = []
					urls_subject = q.pop('imgs_enunciado_urls')
					for i, url in enumerate(urls_subject):
						ref_placeholder = f"[[REF_IMG_{i+1}]]"
						path = download_img(url, folder_q, f"enunciado_{i+1}")
						path_final = path if path else url
						local_imgs_subject.append(path_final)
						q['enunciado'] = q['enunciado'].replace(ref_placeholder, f"{{{path_final}}}")

					urls_alt = q.pop('alternativas_imgs_urls')
					local_imgs_alt = {}
					for l, url in urls_alt.items():
						path = download_img(url, folder_q, f"alt_{l}")
						path_final = path if path else url
						local_imgs_alt[l] = path_final
						original_text = q['alternativas_texto'][l]
						q['alternativas_texto'][l] = f"{original_text} {{{path_final}}}".strip()

					exam_name = next((t for t in exam_types if t in q['tags']), "ENEM")
					ignore = exam_types + ["FUVEST", "UERJ", year_str]
					subjects = [t for t in q['tags'] if not t.isdigit() and t not in ignore]
					main_subject = subjects[0] if subjects else "Geral"

					final_format = {
						"id": q['id'],
						"enunciado": q['enunciado'],
						"alternativas": q['alternativas_texto'],
						"dificuldade": q['dificuldade'],
						"gabarito": q['gabarito'],
						"gabarito_comentado": q['gabarito_comentado'],
						"img_enunciado": local_imgs_subject,
						"img_alternativas": local_imgs_alt,
						"tags": subjects
					}

					if year_str not in organized_data:
						organized_data[year_str] = {}
					if exam_name not in organized_data[year_str]:
						organized_data[year_str][exam_name] = {}
					if main_subject not in organized_data[year_str][exam_name]:
						organized_data[year_str][exam_name][main_subject] = []
					organized_data[year_str][exam_name][main_subject].append(final_format)
				page += 1
	finally:
		with open("banco_questoes_completo.json", "w", encoding='utf-8') as f:
			json.dump(organized_data, f, indent=4, ensure_ascii=False)
		print(f"JSON Ready!")
		driver.quit()

if __name__ == '__main__':
	try:
		main()
	except Exception as e:
		print(f"Error -> {e}")
