"""
YAKE Keyword Extractor

Extracts keywords from parent chunks for sparse (keyword-based) retrieval.

Configuration - CBSE-optimised (NOT defaults):
- n=2: Bigrams only (not 3) - CBSE science focuses on 2-word concepts
- dedupLim=0.7: Aggressive deduplication (removes H₂O/water/hydrogen oxide variants)
- top=8: Top 8 keywords (not 20) - ~375-token parents have 12-15 concepts max

Note: YAKE score is LOWER = MORE RELEVANT. Top-8 are lowest-score keywords.
"""

import yake
from typing import List, Optional
import structlog

logger = structlog.get_logger()


class YakeKeywordExtractor:
    """YAKE keyword extractor with CBSE-optimized parameters.
    
    Attributes:
        extractor: YAKE KeywordExtractor instance
        n: N-gram size (default: 2 for bigrams)
        dedup_lim: Deduplication threshold (default: 0.7)
        top: Number of keywords to extract (default: 8)
    """
    
    def __init__(
        self,
        language: str = "en",
        n: int = 2,
        dedup_lim: float = 0.7,
        top: int = 8,
        features: Optional[dict] = None,
    ):
        """Initialize YAKE extractor.
        
        Args:
            language: Language code (default: "en")
            n: N-gram size (default: 2 for bigrams)
            dedup_lim: Deduplication threshold (default: 0.7)
            top: Number of keywords to extract (default: 8)
            features: Custom feature set (default: None = use YAKE defaults)
        """
        self.n = n
        self.dedup_lim = dedup_lim
        self.top = top
        
        self.extractor = yake.KeywordExtractor(
            lan=language,
            n=n,
            dedupLim=dedup_lim,
            top=top,
            features=features,
        )
        
        logger.info("YAKE extractor initialized",
                   language=language,
                   n=n,
                   dedup_lim=dedup_lim,
                   top=top)
    
    def extract(self, text: str) -> List[str]:
        """Extract keywords from text.
        
        Args:
            text: Text to extract keywords from
        
        Returns:
            List of keyword strings (ordered by relevance)
        """
        if not text or len(text.strip()) < 10:
            return []
        
        try:
            # YAKE returns list of (keyword, score) tuples
            # Score is LOWER = MORE RELEVANT
            keywords_with_scores = self.extractor.extract_keywords(text)
            
            # Extract just the keyword strings
            keywords = [kw for kw, score in keywords_with_scores]
            
            # Validate: all keywords should be bigrams or unigrams
            invalid_keywords = [kw for kw in keywords if len(kw.split()) > 2]
            if invalid_keywords:
                logger.warning("Invalid keywords detected",
                              invalid=invalid_keywords,
                              n=self.n)
                keywords = [kw for kw in keywords if len(kw.split()) <= 2]
            
            logger.debug("Keywords extracted",
                        text_length=len(text),
                        keyword_count=len(keywords),
                        keywords=keywords[:5])  # Log first 5
            
            return keywords
            
        except Exception as e:
            logger.error("Keyword extraction failed",
                        error=str(e),
                        text_length=len(text))
            return []
    
    def extract_with_scores(self, text: str) -> List[tuple]:
        """Extract keywords with relevance scores.
        
        Args:
            text: Text to extract keywords from
        
        Returns:
            List of (keyword, score) tuples
        """
        if not text or len(text.strip()) < 10:
            return []
        
        try:
            return self.extractor.extract_keywords(text)
        except Exception as e:
            logger.error("Keyword extraction failed",
                        error=str(e))
            return []


# Module-level extractor instance (singleton pattern)
_default_extractor: Optional[YakeKeywordExtractor] = None


def get_extractor() -> YakeKeywordExtractor:
    """Get or create default extractor instance.
    
    Returns:
        YakeKeywordExtractor instance
    """
    global _default_extractor
    
    if _default_extractor is None:
        _default_extractor = YakeKeywordExtractor()
    
    return _default_extractor


def extract_keywords(
    text: str,
    n: int = 2,
    dedup_lim: float = 0.7,
    top: int = 8,
) -> List[str]:
    """Extract keywords from text using YAKE.
    
    Convenience function that uses the default extractor.
    
    Args:
        text: Text to extract keywords from
        n: N-gram size (default: 2)
        dedup_lim: Deduplication threshold (default: 0.7)
        top: Number of keywords (default: 8)
    
    Returns:
        List of keyword strings
    
    Examples:
        >>> keywords = extract_keywords("The cell membrane controls transport")
        >>> "cell membrane" in keywords
        True
    """
    if not text or len(text.strip()) < 10:
        return []
    
    # Use default extractor if parameters match
    if n == 2 and dedup_lim == 0.7 and top == 8:
        extractor = get_extractor()
        return extractor.extract(text)
    
    # Create custom extractor for different parameters
    custom_extractor = YakeKeywordExtractor(
        n=n,
        dedup_lim=dedup_lim,
        top=top,
    )
    return custom_extractor.extract(text)


def validate_keywords(
    keywords: List[str],
    expected_min: int = 6,
    expected_max: int = 8,
) -> dict:
    """Validate extracted keywords.
    
    Args:
        keywords: List of keywords
        expected_min: Expected minimum count
        expected_max: Expected maximum count
    
    Returns:
        Dict with validation results
    """
    # Check count
    count_valid = expected_min <= len(keywords) <= expected_max
    
    # Check all are bigrams or unigrams
    length_valid = all(len(kw.split()) <= 2 for kw in keywords)
    
    # Check for duplicates
    unique_count = len(set(keywords))
    no_duplicates = unique_count == len(keywords)
    
    # Check for empty strings
    no_empty = all(kw.strip() for kw in keywords)
    
    return {
        "keyword_count": len(keywords),
        "unique_count": unique_count,
        "count_valid": count_valid,
        "length_valid": length_valid,
        "no_duplicates": no_duplicates,
        "no_empty": no_empty,
        "valid": count_valid and length_valid and no_duplicates and no_empty,
    }


# Example usage:
if __name__ == "__main__":
    test_texts = [
        "The cell membrane is a biological membrane that separates the "
        "interior of all cells from the outside environment. It controls "
        "what enters and exits the cell through selective permeability.",
        
        "Photosynthesis is the process by which green plants and some "
        "other organisms use sunlight to synthesize foods with the help "
        "of chlorophyll. Photosynthesis converts carbon dioxide and water "
        "into glucose and oxygen.",
        
        "Water (H₂O) is a transparent, tasteless, odorless, and nearly "
        "colorless chemical substance. It is essential for all known forms "
        "of life, even though it provides no calories or organic nutrients.",
    ]
    
    extractor = YakeKeywordExtractor()
    
    for i, text in enumerate(test_texts, 1):
        print(f"\n--- Text {i} ---")
        keywords = extractor.extract(text)
        print(f"Keywords ({len(keywords)}): {keywords}")
        
        validation = validate_keywords(keywords)
        print(f"Validation: {validation}")
