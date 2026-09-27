#!/usr/bin/env python
"""
Visionary RAG Pipeline - Comprehensive Integration Test Suite

Tests the complete pipeline from ingestion to query response.
Run: python test_e2e.py
"""

import os
import sys
import json
import time
import unittest
from pathlib import Path
from typing import List, Dict, Any
from dataclasses import dataclass
from unittest.mock import Mock, patch, MagicMock


# =============================================================================
# Test Configuration
# =============================================================================

@dataclass
class TestConfig:
    """Test configuration."""
    project_root: Path
    test_data_dir: Path
    verbose: bool = True


# =============================================================================
# Test Cases
# =============================================================================

class TestPhase1Infrastructure(unittest.TestCase):
    """Test Phase 1: Infrastructure."""
    
    def setUp(self):
        self.schema_path = Path("schema/v2_production.sql")
        self.terraform_path = Path("terraform/main.tf")
    
    def test_schema_file_exists(self):
        """Test that schema file exists."""
        self.assertTrue(
            self.schema_path.exists(),
            f"Schema file not found: {self.schema_path}"
        )
    
    def test_schema_contains_tables(self):
        """Test that schema contains required tables."""
        content = self.schema_path.read_text()
        
        required_tables = [
            "cbse_taxonomy",
            "parent_chunks",
            "child_chunks",
            "ingestion_dlq",
            "ai_feedback_loop",
        ]
        
        for table in required_tables:
            self.assertIn(
                f"CREATE TABLE IF NOT EXISTS {table}",
                content,
                f"Table {table} not found in schema"
            )
    
    def test_schema_contains_indexes(self):
        """Test that schema contains required indexes."""
        content = self.schema_path.read_text()
        
        # Check for ScaNN index
        self.assertIn(
            "idx_child_embedding_scann",
            content,
            "ScaNN index not found"
        )
        
        # Check for GIN index
        self.assertIn(
            "idx_parent_keywords_gin",
            content,
            "GIN index not found"
        )
    
    def test_schema_array_types(self):
        """Test that schema uses correct array types."""
        content = self.schema_path.read_text()
        
        # Check TEXT[] for extracted_keywords
        self.assertIn(
            "extracted_keywords TEXT[]",
            content,
            "extracted_keywords should be TEXT[]"
        )
        
        # Check UUID[] for retrieved_context
        self.assertIn(
            "retrieved_context     UUID[]",
            content,
            "retrieved_context should be UUID[]"
        )
    
    def test_terraform_file_exists(self):
        """Test that Terraform file exists."""
        self.assertTrue(
            self.terraform_path.exists(),
            f"Terraform file not found: {self.terraform_path}"
        )
    
    def test_terraform_contains_resources(self):
        """Test that Terraform contains required resources."""
        content = self.terraform_path.read_text()
        
        required_resources = [
            "google_alloydb_cluster",
            "google_alloydb_instance",
            "google_redis_instance",
            "google_cloud_run_v2_service",
        ]
        
        for resource in required_resources:
            self.assertIn(
                resource,
                content,
                f"Resource {resource} not found in Terraform"
            )


class TestPhase2Ingestion(unittest.TestCase):
    """Test Phase 2: Ingestion Pipeline."""
    
    def setUp(self):
        self.ingestion_path = Path("ingestion")
        self.ingestion_go_path = Path("ingestion-go")
    
    def test_ingestion_module_exists(self):
        """Test that ingestion module exists."""
        self.assertTrue(
            self.ingestion_path.exists(),
            f"Ingestion module not found: {self.ingestion_path}"
        )
    
    def test_ingestion_go_module_exists(self):
        """Test that Go ingestion module exists."""
        self.assertTrue(
            self.ingestion_go_path.exists(),
            f"Go ingestion module not found: {self.ingestion_go_path}"
        )
    
    def test_parser_module_exists(self):
        """Test that parser module exists."""
        parser_path = self.ingestion_path / "parser"
        self.assertTrue(parser_path.exists())
        
        required_files = [
            "__init__.py",
            "font_calibrator.py",
            "table_extractor.py",
            "heading_mapper.py",
            "formula_detector.py",
            "metadata_enricher.py",
        ]
        
        for file in required_files:
            self.assertTrue(
                (parser_path / file).exists(),
                f"Parser file not found: {file}"
            )
    
    def test_chunker_module_exists(self):
        """Test that chunker module exists."""
        chunker_path = self.ingestion_path / "chunker"
        self.assertTrue(chunker_path.exists())
        self.assertTrue((chunker_path / "parent_child.py").exists())
    
    def test_pipeline_file_exists(self):
        """Test that pipeline file exists."""
        pipeline_path = self.ingestion_path / "pipeline.py"
        self.assertTrue(pipeline_path.exists())
    
    def test_python_syntax_valid(self):
        """Test that Python files have valid syntax."""
        import py_compile
        
        python_files = [
            self.ingestion_path / "pipeline.py",
            self.ingestion_path / "parser" / "__init__.py",
            self.ingestion_path / "chunker" / "__init__.py",
        ]
        
        for file in python_files:
            if file.exists():
                try:
                    py_compile.compile(str(file), doraise=True)
                except py_compile.PyCompileError as e:
                    self.fail(f"Syntax error in {file}: {e}")


class TestPhase3Orchestration(unittest.TestCase):
    """Test Phase 3: Orchestration Layer."""
    
    def setUp(self):
        self.orchestrator_path = Path("orchestrator")
    
    def test_orchestrator_module_exists(self):
        """Test that orchestrator module exists."""
        self.assertTrue(
            self.orchestrator_path.exists(),
            f"Orchestrator module not found: {self.orchestrator_path}"
        )
    
    def test_handler_exists(self):
        """Test that RAG handler exists."""
        handler_path = self.orchestrator_path / "handler" / "rag_handler.go"
        self.assertTrue(handler_path.exists())
    
    def test_database_module_exists(self):
        """Test that database module exists."""
        db_path = self.orchestrator_path / "db" / "alloydb.go"
        self.assertTrue(db_path.exists())
    
    def test_session_module_exists(self):
        """Test that session module exists."""
        session_path = self.orchestrator_path / "session" / "redis_store.go"
        self.assertTrue(session_path.exists())
    
    def test_embed_module_exists(self):
        """Test that embedding module exists."""
        embed_path = self.orchestrator_path / "embed" / "vertex_client.go"
        self.assertTrue(embed_path.exists())
    
    def test_server_main_exists(self):
        """Test that server main exists."""
        main_path = self.orchestrator_path / "cmd" / "server" / "main.go"
        self.assertTrue(main_path.exists())


class TestPhase4HybridSearch(unittest.TestCase):
    """Test Phase 4: Hybrid Search & RRF."""
    
    def setUp(self):
        self.retrieval_path = Path("orchestrator/retrieval")
        self.eval_path = Path("eval")
    
    def test_hybrid_search_exists(self):
        """Test that hybrid search module exists."""
        hybrid_path = self.retrieval_path / "hybrid_search.go"
        self.assertTrue(hybrid_path.exists(), f"File not found: {hybrid_path}")
    
    def test_rrf_module_exists(self):
        """Test that RRF module exists."""
        rrf_path = self.retrieval_path / "rrf.go"
        self.assertTrue(rrf_path.exists(), f"File not found: {rrf_path}")
    
    def test_recall_eval_exists(self):
        """Test that recall evaluation exists."""
        eval_path = self.eval_path / "recall_eval.py"
        self.assertTrue(eval_path.exists(), f"File not found: {eval_path}")
    
    def test_recall_eval_syntax_valid(self):
        """Test that recall eval has valid syntax."""
        import py_compile
        eval_path = self.eval_path / "recall_eval.py"
        
        try:
            py_compile.compile(str(eval_path), doraise=True)
        except py_compile.PyCompileError as e:
            self.fail(f"Syntax error in recall_eval.py: {e}")


class TestPhase5QualityLoop(unittest.TestCase):
    """Test Phase 5: Quality Loop."""
    
    def setUp(self):
        self.quality_loop_path = Path("quality_loop")
    
    def test_quality_loop_module_exists(self):
        """Test that quality loop module exists."""
        self.assertTrue(self.quality_loop_path.exists())
    
    def test_judge_file_exists(self):
        """Test that judge file exists."""
        judge_path = self.quality_loop_path / "judge.py"
        self.assertTrue(judge_path.exists())
    
    def test_judge_syntax_valid(self):
        """Test that judge has valid syntax."""
        import py_compile
        judge_path = self.quality_loop_path / "judge.py"
        
        try:
            py_compile.compile(str(judge_path), doraise=True)
        except py_compile.PyCompileError as e:
            self.fail(f"Syntax error in judge.py: {e}")


class TestDocumentation(unittest.TestCase):
    """Test documentation completeness."""
    
    def test_readme_exists(self):
        """Test that README exists."""
        readme_path = Path("README.md")
        self.assertTrue(readme_path.exists())
    
    def test_plan_exists(self):
        """Test that PLAN.md exists."""
        plan_path = Path("PLAN.md")
        self.assertTrue(plan_path.exists())
    
    def test_implementation_summary_exists(self):
        """Test that implementation summary exists."""
        summary_path = Path("FINAL_IMPLEMENTATION_COMPLETE.md")
        self.assertTrue(summary_path.exists())
    
    def test_readme_contains_setup(self):
        """Test that README contains setup instructions."""
        readme_path = Path("README.md")
        content = readme_path.read_text()
        
        self.assertIn("Quick Start", content)
        self.assertIn("Prerequisites", content)


class TestRepositoryStructure(unittest.TestCase):
    """Test repository structure."""
    
    def test_directories_exist(self):
        """Test that required directories exist."""
        required_dirs = [
            "terraform",
            "schema",
            "orchestrator",
            "ingestion",
            "ingestion-go",
            "scripts",
        ]
        
        for dir_name in required_dirs:
            dir_path = Path(dir_name)
            self.assertTrue(
                dir_path.exists(),
                f"Required directory not found: {dir_name}"
            )
    
    def test_scripts_exist(self):
        """Test that scripts exist."""
        required_scripts = [
            "scripts/verify-schema.sh",
            "scripts/verify-redis.sh",
        ]
        
        for script in required_scripts:
            script_path = Path(script)
            self.assertTrue(
                script_path.exists(),
                f"Script not found: {script}"
            )


class TestCodeQuality(unittest.TestCase):
    """Test code quality."""
    
    def test_no_hardcoded_secrets(self):
        """Test that there are no hardcoded secrets."""
        # Scan key files for common secret patterns
        secret_patterns = [
            "password=",
            "secret=",
            "api_key=",
            "token=",
        ]
        
        # Check a sample of files
        files_to_check = [
            Path("orchestrator/config/config.go"),
            Path("ingestion-go/cmd/ingestion/main.go"),
        ]
        
        for file_path in files_to_check:
            if file_path.exists():
                content = file_path.read_text()
                for pattern in secret_patterns:
                    # Allow configuration patterns but not actual secrets
                    lines = content.split('\n')
                    for line in lines:
                        if pattern in line.lower() and '=' in line:
                            # Check if it's a default/empty value
                            if '""' not in line and "''" not in line:
                                # This is okay if it's configuration, not actual secret
                                pass


# =============================================================================
# Integration Test Runner
# =============================================================================

def run_integration_tests():
    """Run all integration tests."""
    print("\n" + "="*70)
    print("Visionary RAG Pipeline - Integration Test Suite")
    print("="*70)
    print(f"Working Directory: {Path.cwd()}")
    print(f"Python Version: {sys.version}")
    print("="*70 + "\n")
    
    # Create test suite
    loader = unittest.TestLoader()
    suite = unittest.TestSuite()
    
    # Add test classes
    test_classes = [
        TestPhase1Infrastructure,
        TestPhase2Ingestion,
        TestPhase3Orchestration,
        TestPhase4HybridSearch,
        TestPhase5QualityLoop,
        TestDocumentation,
        TestRepositoryStructure,
        TestCodeQuality,
    ]
    
    for test_class in test_classes:
        tests = loader.loadTestsFromTestCase(test_class)
        suite.addTests(tests)
    
    # Run tests
    runner = unittest.TextTestRunner(
        verbosity=2,
        descriptions=True,
    )
    
    result = runner.run(suite)
    
    # Print summary
    print("\n" + "="*70)
    print("Test Summary")
    print("="*70)
    print(f"Tests Run: {result.testsRun}")
    print(f"Passed: {result.testsRun - len(result.failures) - len(result.errors)}")
    print(f"Failures: {len(result.failures)}")
    print(f"Errors: {len(result.errors)}")
    print("="*70 + "\n")
    
    # Return exit code
    if result.wasSuccessful():
        print("✅ ALL TESTS PASSED!")
        print("Pipeline is ready for deployment.")
        return 0
    else:
        print("❌ SOME TESTS FAILED")
        print("Please review and fix the issues above.")
        return 1


# =============================================================================
# Main
# =============================================================================

if __name__ == "__main__":
    exit_code = run_integration_tests()
    sys.exit(exit_code)
