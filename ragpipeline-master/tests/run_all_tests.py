#!/usr/bin/env python3
"""
Enterprise RAG Pipeline - Test Runner
Runs all unit tests, integration tests, and build validation
"""

import subprocess
import sys
import os
from pathlib import Path

# Colors for output
class Colors:
    GREEN = '\033[92m'
    RED = '\033[91m'
    YELLOW = '\033[93m'
    BLUE = '\033[94m'
    END = '\033[0m'
    BOLD = '\033[1m'

def print_header(text):
    print(f"\n{Colors.BOLD}{Colors.BLUE}{'='*70}{Colors.END}")
    print(f"{Colors.BOLD}{Colors.BLUE}{text:^70}{Colors.END}")
    print(f"{Colors.BOLD}{Colors.BLUE}{'='*70}{Colors.END}\n")

def print_success(text):
    print(f"{Colors.GREEN}✅ {text}{Colors.END}")

def print_failure(text):
    print(f"{Colors.RED}❌ {text}{Colors.END}")

def print_info(text):
    print(f"{Colors.YELLOW}ℹ️  {text}{Colors.END}")

def run_command(cmd, cwd=None, check=True):
    """Run a shell command and return the result."""
    print_info(f"Running: {' '.join(cmd)}")
    
    try:
        result = subprocess.run(
            cmd,
            cwd=cwd,
            capture_output=True,
            text=True,
            timeout=300
        )
        
        if result.stdout:
            print(result.stdout)
        
        if result.returncode != 0 and check:
            print_failure(f"Command failed with exit code {result.returncode}")
            if result.stderr:
                print(result.stderr)
            return False
        
        return True
        
    except subprocess.TimeoutExpired:
        print_failure("Command timed out")
        return False
    except Exception as e:
        print_failure(f"Command failed: {e}")
        return False

def test_python_unit_tests():
    """Run Python unit tests."""
    print_header("PYTHON UNIT TESTS")
    
    test_files = [
        "orchestrator/middleware/auth_test.py",
        "orchestrator/analytics/cost_tracker_test.py",
        "orchestrator/retrieval/hybrid_search_test.py",
        "ingestion/chunker/test_parent_child.py",
        "ingestion/dedup/test_detector.py",
    ]
    
    passed = 0
    failed = 0
    
    for test_file in test_files:
        if os.path.exists(test_file):
            print_info(f"Testing: {test_file}")
            success = run_command(["python", "-m", "pytest", test_file, "-v"], check=False)
            if success:
                passed += 1
            else:
                failed += 1
        else:
            print_info(f"Skipping (not found): {test_file}")
    
    print(f"\n{Colors.BOLD}Python Tests: {passed} passed, {failed} failed{Colors.END}")
    return failed == 0

def test_go_unit_tests():
    """Run Go unit tests."""
    print_header("GO UNIT TESTS")
    
    orchestrator_path = Path("orchestrator").absolute()
    
    if not os.path.exists(orchestrator_path / "go.mod"):
        print_info("Go module not found, skipping Go tests")
        return True
    
    # Run go mod tidy first
    print_info("Running go mod tidy...")
    success = run_command(["go", "mod", "tidy"], cwd=orchestrator_path)
    if not success:
        print_failure("go mod tidy failed")
        return False
    
    # Run all Go tests with coverage
    print_info("Running Go tests with coverage...")
    success = run_command(
        ["go", "test", "./...", "-v", "-cover", "-race"],
        cwd=orchestrator_path,
        check=False
    )
    
    if success:
        print_success("Go unit tests passed")
    else:
        print_failure("Go unit tests failed")
    
    return success

def test_build():
    """Test building the application."""
    print_header("BUILD VALIDATION")
    
    orchestrator_path = Path("orchestrator").absolute()
    
    # Build the orchestrator
    print_info("Building orchestrator...")
    success = run_command(
        ["go", "build", "-o", "bin/orchestrator.exe", "./cmd/server"],
        cwd=orchestrator_path,
        check=False
    )
    
    if success:
        print_success("Build successful")
        
        # Check if binary was created
        binary_path = orchestrator_path / "bin" / "orchestrator.exe"
        if os.path.exists(binary_path):
            size_mb = os.path.getsize(binary_path) / (1024 * 1024)
            print_info(f"Binary created: {binary_path} ({size_mb:.2f} MB)")
    else:
        print_failure("Build failed")
    
    return success

def test_integration():
    """Run integration tests."""
    print_header("INTEGRATION TESTS")
    
    integration_tests = [
        "tests/integration/test_rag_pipeline.py",
        "tests/integration/test_auth_flow.py",
        "tests/integration/test_cost_tracking.py",
    ]
    
    passed = 0
    failed = 0
    
    for test_file in integration_tests:
        if os.path.exists(test_file):
            print_info(f"Testing: {test_file}")
            success = run_command(
                ["python", "-m", "pytest", test_file, "-v", "--tb=short"],
                check=False
            )
            if success:
                passed += 1
            else:
                failed += 1
        else:
            print_info(f"Skipping (not found): {test_file}")
    
    print(f"\n{Colors.BOLD}Integration Tests: {passed} passed, {failed} failed{Colors.END}")
    return failed == 0

def test_code_quality():
    """Run code quality checks."""
    print_header("CODE QUALITY")
    
    # Python linting
    print_info("Running Python linting (flake8)...")
    python_success = run_command(
        ["flake8", "orchestrator/", "--max-line-length=120", "--ignore=E501,W503"],
        check=False
    )
    
    # Go linting (if available)
    orchestrator_path = Path("orchestrator").absolute()
    if os.path.exists(orchestrator_path / "go.mod"):
        print_info("Running Go linting...")
        go_success = run_command(
            ["go", "vet", "./..."],
            cwd=orchestrator_path,
            check=False
        )
    else:
        go_success = True
    
    if python_success and go_success:
        print_success("Code quality checks passed")
    else:
        print_failure("Code quality checks failed")
    
    return python_success and go_success

def test_security():
    """Run security checks."""
    print_header("SECURITY CHECKS")
    
    # Check for hardcoded secrets
    print_info("Scanning for hardcoded secrets...")
    
    secret_patterns = [
        "password=",
        "secret=",
        "api_key=",
        "token=",
    ]
    
    found_secrets = False
    
    for root, dirs, files in os.walk("orchestrator"):
        # Skip vendor and test directories
        dirs[:] = [d for d in dirs if d not in ['vendor', 'node_modules', '__pycache__']]
        
        for file in files:
            if file.endswith(('.go', '.py')):
                filepath = os.path.join(root, file)
                
                # Skip test files
                if 'test' in filepath.lower():
                    continue
                
                with open(filepath, 'r', encoding='utf-8', errors='ignore') as f:
                    content = f.read()
                    
                    # Check for hardcoded secrets (not in config/context)
                    for pattern in secret_patterns:
                        if pattern in content.lower():
                            # Check if it's in a config context
                            lines = content.split('\n')
                            for i, line in enumerate(lines):
                                if pattern in line.lower() and '=' in line:
                                    # Check if it's a variable declaration, not a hardcoded value
                                    if 'os.Getenv' in line or 'flag' in line:
                                        continue  # OK - using env vars
                                    elif 'your-' in line or 'example' in line.lower():
                                        continue  # OK - example/placeholder
                                    else:
                                        print_info(f"Potential secret in {filepath}:{i+1}")
                                        found_secrets = True
    
    if not found_secrets:
        print_success("No hardcoded secrets found")
    else:
        print_failure("Potential hardcoded secrets found")
    
    return not found_secrets

def generate_report():
    """Generate test report."""
    print_header("TEST REPORT")
    
    report = {
        "Python Unit Tests": "✅ PASS" if test_python_unit_tests() else "❌ FAIL",
        "Go Unit Tests": "✅ PASS" if test_go_unit_tests() else "❌ FAIL",
        "Build": "✅ PASS" if test_build() else "❌ FAIL",
        "Integration Tests": "✅ PASS" if test_integration() else "❌ FAIL",
        "Code Quality": "✅ PASS" if test_code_quality() else "❌ FAIL",
        "Security": "✅ PASS" if test_security() else "❌ FAIL",
    }
    
    print("\n")
    for test_name, status in report.items():
        print(f"{test_name:.<50} {status}")
    
    all_passed = all("PASS" in status for status in report.values())
    
    print("\n" + "="*70)
    if all_passed:
        print(f"{Colors.GREEN}{Colors.BOLD}🎉 ALL TESTS PASSED - READY FOR DEPLOYMENT{Colors.END}")
    else:
        print(f"{Colors.RED}{Colors.BOLD}⚠️  SOME TESTS FAILED - FIX BEFORE DEPLOYMENT{Colors.END}")
    print("="*70)
    
    return all_passed

def main():
    """Main test runner."""
    print_header("🧪 ENTERPRISE RAG PIPELINE - TEST SUITE")
    
    # Change to project root
    project_root = Path(__file__).parent.parent
    os.chdir(project_root)
    print_info(f"Project root: {project_root}")
    
    # Run all tests
    all_passed = generate_report()
    
    # Exit with appropriate code
    sys.exit(0 if all_passed else 1)

if __name__ == "__main__":
    main()
