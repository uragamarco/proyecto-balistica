#!/usr/bin/env python3
"""
Script de pruebas de rendimiento para el sistema de análisis balístico
Mide métricas específicas de rendimiento y genera un reporte detallado
"""

import requests
import time
import json
import os
import statistics
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path
import psutil
import threading

class PerformanceMonitor:
    def __init__(self):
        self.cpu_usage = []
        self.memory_usage = []
        self.monitoring = False
        self.monitor_thread = None
    
    def start_monitoring(self):
        self.monitoring = True
        self.monitor_thread = threading.Thread(target=self._monitor_resources)
        self.monitor_thread.start()
    
    def stop_monitoring(self):
        self.monitoring = False
        if self.monitor_thread:
            self.monitor_thread.join()
    
    def _monitor_resources(self):
        while self.monitoring:
            self.cpu_usage.append(psutil.cpu_percent())
            self.memory_usage.append(psutil.virtual_memory().percent)
            time.sleep(0.5)
    
    def get_stats(self):
        return {
            'cpu': {
                'avg': statistics.mean(self.cpu_usage) if self.cpu_usage else 0,
                'max': max(self.cpu_usage) if self.cpu_usage else 0,
                'min': min(self.cpu_usage) if self.cpu_usage else 0
            },
            'memory': {
                'avg': statistics.mean(self.memory_usage) if self.memory_usage else 0,
                'max': max(self.memory_usage) if self.memory_usage else 0,
                'min': min(self.memory_usage) if self.memory_usage else 0
            }
        }

class BallisticPerformanceTest:
    def __init__(self, base_url="http://localhost:8080"):
        self.base_url = base_url
        self.test_images_dir = Path("test_images")
        self.results = {
            'individual_analysis': [],
            'comparison_tests': [],
            'concurrent_tests': [],
            'resource_usage': {},
            'summary': {}
        }
        self.monitor = PerformanceMonitor()
    
    def check_server_health(self):
        """Verifica que el servidor esté funcionando"""
        try:
            # Intentar con diferentes endpoints
            endpoints = ["/health", "/api/health", "/"]
            for endpoint in endpoints:
                try:
                    response = requests.get(f"{self.base_url}{endpoint}", timeout=5)
                    if response.status_code in [200, 404]:  # 404 también indica que el servidor responde
                        return True
                except:
                    continue
            return False
        except:
            return False
    
    def test_individual_analysis(self, image_path, iterations=5):
        """Prueba el análisis individual de imágenes"""
        print(f"\n🔍 Probando análisis individual: {image_path.name}")
        
        times = []
        success_count = 0
        
        for i in range(iterations):
            try:
                start_time = time.time()
                
                with open(image_path, 'rb') as f:
                    # Determinar el tipo MIME correcto basado en la extensión
                    if image_path.suffix.lower() == '.jpg' or image_path.suffix.lower() == '.jpeg':
                        mime_type = 'image/jpeg'
                    elif image_path.suffix.lower() == '.png':
                        mime_type = 'image/png'
                    elif image_path.suffix.lower() == '.tiff' or image_path.suffix.lower() == '.tif':
                        mime_type = 'image/tiff'
                    else:
                        mime_type = 'image/jpeg'  # Por defecto
                    
                    files = {'image': (image_path.name, f, mime_type)}
                    response = requests.post(
                        f"{self.base_url}/api/process",
                        files=files,
                        timeout=120
                    )
                
                end_time = time.time()
                duration = end_time - start_time
                
                if response.status_code == 200:
                    success_count += 1
                    times.append(duration)
                    print(f"  ✅ Iteración {i+1}: {duration:.2f}s")
                else:
                    print(f"  ❌ Iteración {i+1}: Error {response.status_code}")
                    
            except Exception as e:
                print(f"  ❌ Iteración {i+1}: {str(e)}")
        
        if times:
            result = {
                'image': image_path.name,
                'size_mb': image_path.stat().st_size / (1024 * 1024),
                'iterations': iterations,
                'success_rate': success_count / iterations * 100,
                'avg_time': statistics.mean(times),
                'min_time': min(times),
                'max_time': max(times),
                'std_dev': statistics.stdev(times) if len(times) > 1 else 0
            }
            self.results['individual_analysis'].append(result)
            return result
        
        return None
    
    def test_concurrent_analysis(self, image_paths, concurrent_users=3):
        """Prueba análisis concurrente con múltiples usuarios"""
        print(f"\n🚀 Probando análisis concurrente con {concurrent_users} usuarios")
        
        def analyze_image(image_path):
            try:
                start_time = time.time()
                with open(image_path, 'rb') as f:
                    # Determinar el tipo MIME correcto basado en la extensión
                    if image_path.suffix.lower() == '.jpg' or image_path.suffix.lower() == '.jpeg':
                        mime_type = 'image/jpeg'
                    elif image_path.suffix.lower() == '.png':
                        mime_type = 'image/png'
                    elif image_path.suffix.lower() == '.tiff' or image_path.suffix.lower() == '.tif':
                        mime_type = 'image/tiff'
                    else:
                        mime_type = 'image/jpeg'  # Por defecto
                    
                    files = {'image': (image_path.name, f, mime_type)}
                    response = requests.post(
                        f"{self.base_url}/api/process",
                        files=files,
                        timeout=120
                    )
                end_time = time.time()
                
                return {
                    'image': image_path.name,
                    'duration': end_time - start_time,
                    'success': response.status_code == 200,
                    'status_code': response.status_code
                }
            except Exception as e:
                return {
                    'image': image_path.name,
                    'duration': 0,
                    'success': False,
                    'error': str(e)
                }
        
        start_time = time.time()
        
        with ThreadPoolExecutor(max_workers=concurrent_users) as executor:
            futures = [executor.submit(analyze_image, img) for img in image_paths[:concurrent_users]]
            results = [future.result() for future in as_completed(futures)]
        
        end_time = time.time()
        total_time = end_time - start_time
        
        successful_results = [r for r in results if r['success']]
        
        concurrent_result = {
            'concurrent_users': concurrent_users,
            'total_requests': len(results),
            'successful_requests': len(successful_results),
            'success_rate': len(successful_results) / len(results) * 100,
            'total_time': total_time,
            'avg_response_time': statistics.mean([r['duration'] for r in successful_results]) if successful_results else 0,
            'throughput': len(successful_results) / total_time if total_time > 0 else 0
        }
        
        self.results['concurrent_tests'].append(concurrent_result)
        print(f"  📊 Tasa de éxito: {concurrent_result['success_rate']:.1f}%")
        print(f"  ⚡ Throughput: {concurrent_result['throughput']:.2f} req/s")
        
        return concurrent_result
    
    def run_comprehensive_test(self):
        """Ejecuta todas las pruebas de rendimiento"""
        print("🎯 Iniciando pruebas de rendimiento del sistema balístico")
        print("=" * 60)
        
        # Verificar servidor
        if not self.check_server_health():
            print("❌ El servidor no está disponible")
            return
        
        print("✅ Servidor disponible")
        
        # Buscar imágenes de prueba
        test_images = list(self.test_images_dir.glob("*.jpg")) + \
                     list(self.test_images_dir.glob("*.png")) + \
                     list(self.test_images_dir.glob("*.tiff"))
        
        if not test_images:
            print("❌ No se encontraron imágenes de prueba")
            return
        
        print(f"📁 Encontradas {len(test_images)} imágenes de prueba")
        
        # Iniciar monitoreo de recursos
        self.monitor.start_monitoring()
        
        try:
            # Pruebas individuales
            print("\n📋 PRUEBAS DE ANÁLISIS INDIVIDUAL")
            print("-" * 40)
            
            for image_path in test_images[:3]:  # Limitar a 3 imágenes
                self.test_individual_analysis(image_path, iterations=3)
            
            # Pruebas concurrentes
            print("\n📋 PRUEBAS DE ANÁLISIS CONCURRENTE")
            print("-" * 40)
            
            if len(test_images) >= 2:
                self.test_concurrent_analysis(test_images, concurrent_users=2)
                self.test_concurrent_analysis(test_images, concurrent_users=3)
        
        finally:
            # Detener monitoreo
            self.monitor.stop_monitoring()
            self.results['resource_usage'] = self.monitor.get_stats()
        
        # Generar resumen
        self.generate_summary()
        
        # Guardar resultados
        self.save_results()
        
        # Mostrar reporte
        self.print_report()
    
    def generate_summary(self):
        """Genera resumen de las pruebas"""
        individual_tests = self.results['individual_analysis']
        concurrent_tests = self.results['concurrent_tests']
        
        if individual_tests:
            avg_times = [t['avg_time'] for t in individual_tests]
            success_rates = [t['success_rate'] for t in individual_tests]
            
            self.results['summary'] = {
                'individual_analysis': {
                    'avg_processing_time': statistics.mean(avg_times),
                    'max_processing_time': max(avg_times),
                    'min_processing_time': min(avg_times),
                    'overall_success_rate': statistics.mean(success_rates)
                },
                'concurrent_analysis': {
                    'max_throughput': max([t['throughput'] for t in concurrent_tests]) if concurrent_tests else 0,
                    'avg_success_rate': statistics.mean([t['success_rate'] for t in concurrent_tests]) if concurrent_tests else 0
                },
                'resource_usage': self.results['resource_usage']
            }
    
    def save_results(self):
        """Guarda los resultados en un archivo JSON"""
        timestamp = time.strftime("%Y%m%d_%H%M%S")
        filename = f"performance_results_{timestamp}.json"
        
        with open(filename, 'w') as f:
            json.dump(self.results, f, indent=2)
        
        print(f"\n💾 Resultados guardados en: {filename}")
    
    def print_report(self):
        """Imprime el reporte de rendimiento"""
        print("\n" + "=" * 60)
        print("📊 REPORTE DE RENDIMIENTO")
        print("=" * 60)
        
        summary = self.results['summary']
        
        if 'individual_analysis' in summary:
            ind = summary['individual_analysis']
            print(f"\n🔍 ANÁLISIS INDIVIDUAL:")
            print(f"  • Tiempo promedio: {ind['avg_processing_time']:.2f}s")
            print(f"  • Tiempo mínimo: {ind['min_processing_time']:.2f}s")
            print(f"  • Tiempo máximo: {ind['max_processing_time']:.2f}s")
            print(f"  • Tasa de éxito: {ind['overall_success_rate']:.1f}%")
        
        if 'concurrent_analysis' in summary:
            conc = summary['concurrent_analysis']
            print(f"\n🚀 ANÁLISIS CONCURRENTE:")
            print(f"  • Throughput máximo: {conc['max_throughput']:.2f} req/s")
            print(f"  • Tasa de éxito promedio: {conc['avg_success_rate']:.1f}%")
        
        if 'resource_usage' in summary:
            res = summary['resource_usage']
            print(f"\n💻 USO DE RECURSOS:")
            print(f"  • CPU promedio: {res['cpu']['avg']:.1f}%")
            print(f"  • CPU máximo: {res['cpu']['max']:.1f}%")
            print(f"  • Memoria promedio: {res['memory']['avg']:.1f}%")
            print(f"  • Memoria máxima: {res['memory']['max']:.1f}%")
        
        # Recomendaciones
        print(f"\n💡 RECOMENDACIONES:")
        if summary.get('individual_analysis', {}).get('avg_processing_time', 0) > 10:
            print("  ⚠️  Tiempo de procesamiento alto (>10s) - Optimizar algoritmos")
        if summary.get('resource_usage', {}).get('cpu', {}).get('avg', 0) > 80:
            print("  ⚠️  Alto uso de CPU - Considerar optimización o escalado")
        if summary.get('resource_usage', {}).get('memory', {}).get('avg', 0) > 80:
            print("  ⚠️  Alto uso de memoria - Revisar gestión de memoria")
        if summary.get('concurrent_analysis', {}).get('avg_success_rate', 100) < 95:
            print("  ⚠️  Baja tasa de éxito en concurrencia - Revisar manejo de carga")

if __name__ == "__main__":
    # Crear directorio de imágenes de prueba si no existe
    test_dir = Path("test_images")
    test_dir.mkdir(exist_ok=True)
    
    # Ejecutar pruebas
    tester = BallisticPerformanceTest()
    tester.run_comprehensive_test()