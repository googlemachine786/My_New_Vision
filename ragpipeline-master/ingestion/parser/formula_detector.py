"""
Pass 4: Formula Detector

Identifies chemical formulas, mathematical equations, and scientific notation in text.

Inputs:  str (prose text), span flags from PyMuPDF
Outputs: FormulaResult with content_type and annotations

Patterns:
  - chemical_formula: H₂O, CO₂, NaCl, etc. (with subscript detection)
  - subscript_flag: PyMuPDF flags & 1 or 2 (superscript/subscript)
  - measurement: 10.5 km, 100 m/s, 25°C, etc.
  - equation: E = mc², F = ma, etc.

Content types:
  - "prose": Regular text
  - "formula": Contains chemical/mathematical formulas
  - "equation": Contains mathematical equations
"""

import re
from dataclasses import dataclass, field
from typing import List, Optional, Tuple
import structlog

logger = structlog.get_logger()


@dataclass
class FormulaResult:
    """Formula detection result.
    
    Attributes:
        content_type: "prose", "formula", or "equation"
        formula_annotations: List of detected formulas
        has_subscripts: True if subscript/superscript detected
        has_measurements: True if measurements detected
        confidence: Confidence score (0.0-1.0)
    """
    content_type: str = "prose"
    formula_annotations: List[str] = field(default_factory=list)
    has_subscripts: bool = False
    has_measurements: bool = False
    confidence: float = 0.0
    
    def is_formula(self) -> bool:
        """Check if content is formula-type."""
        return self.content_type == "formula"
    
    def is_equation(self) -> bool:
        """Check if content is equation-type."""
        return self.content_type == "equation"


# Chemical formula pattern
# Matches: H₂O, CO₂, NaCl, C₆H₁₂O₆, etc.
CHEMICAL_FORMULA_REGEX = re.compile(
    r"\b[A-Z][a-z]?\d*[₀₁₂₃₄₅₆₇₈₉]*"
    r"(?:[A-Z][a-z]?\d*[₀₁₂₃₄₅₆₇₈₉]*)+\b"
)

# Subscript/superscript Unicode characters
SUBSCRIPT_CHARS = "₀₁₂₃₄₅₆₇₈₉"
SUPERSCRIPT_CHARS = "⁰¹²³⁴⁵⁶⁷⁸⁹"

# Measurement pattern
# Matches: 10.5 km, 100 m/s, 25°C, 3.0×10⁸ m/s, etc.
MEASUREMENT_REGEX = re.compile(
    r"\d+(?:\.\d+)?\s*"
    r"(?:km|m|cm|mm|kg|g|mg|°C|K|J|W|Hz|nm|μm|mol|L|ml|s|ms|N|Pa|V|A|Ω|T|Hz|eV|u)"
)

# Equation pattern
# Matches: E = mc², F = ma, PV = nRT, etc.
EQUATION_REGEX = re.compile(
    r"[A-Z][a-z]?\s*[=≠≤≥]\s*[A-Z0-9][A-Za-z0-9\s\+\-\*\/\(\)]{2,}"
)

# Scientific notation
SCIENTIFIC_NOTATION_REGEX = re.compile(
    r"\d+(?:\.\d+)?\s*×\s*10[⁺⁻]?[⁰¹²³⁴⁵⁶⁷⁸⁹]+"
)


def detect_formulas(
    text: str,
    span_flags: int = 0
) -> FormulaResult:
    """Detect formulas, equations, and measurements in text.
    
    Args:
        text: Text to analyze
        span_flags: PyMuPDF span flags (for subscript/superscript detection)
    
    Returns:
        FormulaResult with detection results
    """
    annotations: List[str] = []
    has_subscripts = False
    has_measurements = False
    has_equations = False
    has_chemical = False
    
    # Check for subscript/superscript flags
    # PyMuPDF flags: bit 0 = superscript, bit 1 = subscript
    if span_flags & 0x01 or span_flags & 0x02:
        has_subscripts = True
        logger.debug("Subscript/superscript flag detected", flags=span_flags)
    
    # Check for Unicode subscripts/superscripts in text
    if any(c in text for c in SUBSCRIPT_CHARS + SUPERSCRIPT_CHARS):
        has_subscripts = True
    
    # Find chemical formulas
    chemical_matches = CHEMICAL_FORMULA_REGEX.findall(text)
    if chemical_matches:
        has_chemical = True
        annotations.extend(chemical_matches)
        logger.debug("Chemical formulas found", formulas=chemical_matches)
    
    # Find measurements
    measurement_matches = MEASUREMENT_REGEX.findall(text)
    if measurement_matches:
        has_measurements = True
        annotations.extend(measurement_matches)
        logger.debug("Measurements found", measurements=measurement_matches)
    
    # Find equations
    equation_matches = EQUATION_REGEX.findall(text)
    if equation_matches:
        has_equations = True
        annotations.extend(equation_matches)
        logger.debug("Equations found", equations=equation_matches)
    
    # Find scientific notation
    scientific_matches = SCIENTIFIC_NOTATION_REGEX.findall(text)
    if scientific_matches:
        has_subscripts = True  # Scientific notation uses superscripts
        annotations.extend(scientific_matches)
    
    # Determine content type
    content_type = "prose"
    confidence = 0.0
    
    if has_equations:
        content_type = "equation"
        confidence = 0.9
    elif has_chemical or has_subscripts:
        content_type = "formula"
        confidence = 0.85 if has_subscripts else 0.7
    elif has_measurements:
        # Measurements alone don't make it a formula, but increase confidence
        content_type = "prose"
        confidence = 0.3
    
    # Boost confidence if multiple indicators
    if sum([has_chemical, has_subscripts, has_equations, has_measurements]) > 1:
        confidence = min(confidence + 0.15, 1.0)
    
    return FormulaResult(
        content_type=content_type,
        formula_annotations=annotations,
        has_subscripts=has_subscripts,
        has_measurements=has_measurements,
        confidence=round(confidence, 2),
    )


def detect_formulas_in_text(
    text: str
) -> FormulaResult:
    """Detect formulas without span flags (fallback).
    
    Args:
        text: Text to analyze
    
    Returns:
        FormulaResult with detection results
    """
    return detect_formulas(text, span_flags=0)


def is_formula_heavy(
    text: str,
    threshold: float = 0.3
) -> bool:
    """Check if text is formula-heavy (for special handling).
    
    Formula-heavy text may need special chunking to avoid splitting
    related formulas across chunks.
    
    Args:
        text: Text to analyze
        threshold: Ratio of formula tokens to total tokens
    
    Returns:
        True if text is formula-heavy
    """
    result = detect_formulas_in_text(text)
    
    if not result.formula_annotations:
        return False
    
    # Estimate formula token ratio
    formula_tokens = sum(len(a) for a in result.formula_annotations)
    total_tokens = len(text)
    
    return (formula_tokens / total_tokens) > threshold if total_tokens > 0 else False


def extract_formulas_only(
    text: str
) -> List[str]:
    """Extract only formula content from text.
    
    Args:
        text: Text to extract formulas from
    
    Returns:
        List of unique formulas found
    """
    result = detect_formulas_in_text(text)
    return list(set(result.formula_annotations))


# Example usage:
if __name__ == "__main__":
    test_cases = [
        "Water (H₂O) is essential for life.",
        "The chemical formula for glucose is C₆H₁₂O₆.",
        "E = mc² is Einstein's mass-energy equivalence.",
        "The speed of light is 3.0×10⁸ m/s.",
        "The cell membrane controls what enters and exits the cell.",
        "CO₂ + H₂O → C₆H₁₂O₆ + O₂ (photosynthesis)",
        "Temperature: 25°C, Pressure: 101.3 kPa",
    ]
    
    for text in test_cases:
        result = detect_formulas_in_text(text)
        print(f"\nText: {text}")
        print(f"Type: {result.content_type} (confidence: {result.confidence})")
        if result.formula_annotations:
            print(f"Annotations: {result.formula_annotations}")
