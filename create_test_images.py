#!/usr/bin/env python3
"""
Script para crear imágenes de prueba sintéticas para el análisis balístico
"""

import numpy as np
from PIL import Image, ImageDraw, ImageFilter
import os
from pathlib import Path
import random

def create_test_images():
    """Crea imágenes de prueba sintéticas que simulan casquillos balísticos"""
    
    test_dir = Path("test_images")
    test_dir.mkdir(exist_ok=True)
    
    print("🎨 Creando imágenes de prueba sintéticas...")
    
    # Imagen 1: Casquillo con marcas de percutor circulares
    img1 = Image.new('RGB', (800, 600), color='lightgray')
    draw1 = ImageDraw.Draw(img1)
    
    # Base del casquillo
    draw1.ellipse([200, 150, 600, 450], fill='darkgray', outline='black', width=3)
    
    # Marca de percutor central
    draw1.ellipse([350, 250, 450, 350], fill='gray', outline='darkgray', width=2)
    
    # Marcas radiales (estrías)
    center_x, center_y = 400, 300
    for angle in range(0, 360, 15):
        x1 = center_x + 80 * np.cos(np.radians(angle))
        y1 = center_y + 80 * np.sin(np.radians(angle))
        x2 = center_x + 120 * np.cos(np.radians(angle))
        y2 = center_y + 120 * np.sin(np.radians(angle))
        draw1.line([(x1, y1), (x2, y2)], fill='black', width=1)
    
    # Añadir ruido
    img1 = img1.filter(ImageFilter.GaussianBlur(radius=0.5))
    img1.save(test_dir / "casquillo_circular_800x600.jpg", quality=85)
    
    # Imagen 2: Casquillo con marcas rectangulares
    img2 = Image.new('RGB', (1024, 768), color='lightgray')
    draw2 = ImageDraw.Draw(img2)
    
    # Base del casquillo
    draw2.ellipse([250, 200, 774, 568], fill='darkgray', outline='black', width=4)
    
    # Marca de percutor rectangular
    draw2.rectangle([450, 330, 574, 438], fill='gray', outline='darkgray', width=2)
    
    # Estrías lineales
    for i in range(8):
        y = 250 + i * 35
        draw2.line([(300, y), (724, y)], fill='black', width=2)
    
    # Marcas adicionales
    for i in range(12):
        angle = i * 30
        x1 = 512 + 150 * np.cos(np.radians(angle))
        y1 = 384 + 150 * np.sin(np.radians(angle))
        x2 = 512 + 180 * np.cos(np.radians(angle))
        y2 = 384 + 180 * np.sin(np.radians(angle))
        draw2.line([(x1, y1), (x2, y2)], fill='darkgray', width=3)
    
    img2 = img2.filter(ImageFilter.UnsharpMask())
    img2.save(test_dir / "casquillo_rectangular_1024x768.jpg", quality=90)
    
    # Imagen 3: Casquillo pequeño con alta resolución
    img3 = Image.new('RGB', (512, 512), color='white')
    draw3 = ImageDraw.Draw(img3)
    
    # Base del casquillo
    draw3.ellipse([50, 50, 462, 462], fill='lightgray', outline='black', width=2)
    
    # Marca de percutor compleja
    draw3.ellipse([200, 200, 312, 312], fill='gray', outline='black', width=1)
    draw3.ellipse([220, 220, 292, 292], fill='darkgray')
    
    # Patrón de estrías en espiral
    center_x, center_y = 256, 256
    for i in range(100):
        angle = i * 7.2  # 720 grados en total
        radius = 80 + i * 0.8
        x = center_x + radius * np.cos(np.radians(angle))
        y = center_y + radius * np.sin(np.radians(angle))
        if radius < 180:
            draw3.ellipse([x-1, y-1, x+1, y+1], fill='black')
    
    img3.save(test_dir / "casquillo_pequeno_512x512.png")
    
    # Imagen 4: Imagen grande para pruebas de memoria
    img4 = Image.new('RGB', (2048, 1536), color='lightblue')
    draw4 = ImageDraw.Draw(img4)
    
    # Base del casquillo grande
    draw4.ellipse([400, 300, 1648, 1236], fill='silver', outline='black', width=6)
    
    # Marca de percutor central grande
    draw4.ellipse([900, 650, 1148, 886], fill='gray', outline='darkgray', width=4)
    
    # Muchas marcas pequeñas para simular textura compleja
    for _ in range(500):
        x = random.randint(500, 1548)
        y = random.randint(400, 1136)
        size = random.randint(2, 8)
        draw4.ellipse([x, y, x+size, y+size], fill='darkgray')
    
    # Estrías radiales densas
    center_x, center_y = 1024, 768
    for angle in range(0, 360, 3):
        x1 = center_x + 200 * np.cos(np.radians(angle))
        y1 = center_y + 200 * np.sin(np.radians(angle))
        x2 = center_x + 400 * np.cos(np.radians(angle))
        y2 = center_y + 400 * np.sin(np.radians(angle))
        draw4.line([(x1, y1), (x2, y2)], fill='black', width=2)
    
    img4.save(test_dir / "casquillo_grande_2048x1536.tiff")
    
    # Imagen 5: Imagen con mucho ruido para pruebas de robustez
    img5 = Image.new('RGB', (640, 480), color='gray')
    draw5 = ImageDraw.Draw(img5)
    
    # Base del casquillo
    draw5.ellipse([120, 90, 520, 390], fill='lightgray', outline='black', width=3)
    
    # Marca de percutor
    draw5.ellipse([280, 200, 360, 280], fill='darkgray', outline='black', width=2)
    
    # Añadir mucho ruido
    pixels = list(img5.getdata())
    noisy_pixels = []
    for pixel in pixels:
        r, g, b = pixel
        # Añadir ruido gaussiano
        noise = random.gauss(0, 20)
        r = max(0, min(255, r + noise))
        g = max(0, min(255, g + noise))
        b = max(0, min(255, b + noise))
        noisy_pixels.append((int(r), int(g), int(b)))
    
    img5.putdata(noisy_pixels)
    img5.save(test_dir / "casquillo_ruidoso_640x480.jpg", quality=70)
    
    print(f"✅ Creadas 5 imágenes de prueba en {test_dir}:")
    print("  • casquillo_circular_800x600.jpg (800x600, marcas circulares)")
    print("  • casquillo_rectangular_1024x768.jpg (1024x768, marcas rectangulares)")
    print("  • casquillo_pequeno_512x512.png (512x512, alta resolución)")
    print("  • casquillo_grande_2048x1536.tiff (2048x1536, prueba de memoria)")
    print("  • casquillo_ruidoso_640x480.jpg (640x480, con ruido)")
    
    # Mostrar tamaños de archivo
    print("\n📊 Tamaños de archivo:")
    for img_file in test_dir.glob("*"):
        size_mb = img_file.stat().st_size / (1024 * 1024)
        print(f"  • {img_file.name}: {size_mb:.2f} MB")

if __name__ == "__main__":
    create_test_images()