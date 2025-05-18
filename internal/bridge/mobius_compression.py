#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
WHO: PythonMobiusCompression
WHAT: Python implementation of Möbius compression
WHEN: During data compression operations
WHERE: System Layer 3 (Higher Thought)
WHY: To provide efficient data compression
HOW: Using Möbius compression formula with context-awareness
EXTENT: All Python-managed compression operations
"""

import json
import math
import time
from typing import Any, Dict


class MobiusCompression:
    """
    WHO: PythonMobiusCompression
    WHAT: Python implementation of Möbius compression
    WHEN: During data compression operations
    WHERE: System Layer 3 (Higher Thought)
    WHY: To provide efficient data compression
    HOW: Using Möbius compression formula with context-awareness
    EXTENT: All Python-managed compression operations
    """
    
    # Class variables to track statistics
    _stats = {
        "compressions": 0,
        "decompressions": 0,
        "totalOriginalSize": 0,
        "totalCompressedSize": 0,
        "compressionTime": 0.0,
        "decompressionTime": 0.0,
        "errors": 0,
        "lastRun": 0
    }
    
    # Default compression factors
    DEFAULT_FACTORS = {
        "B": 1.5,  # Base factor for who/what dimensions
        "V": 2.3,  # Variance factor for where/how dimensions
        "I": 1.2,  # Intent factor for why dimension
        "G": 0.8,  # Temporal gradient for when dimension
        "F": 1.1,  # Fidelity factor for extent dimension
        "E": 0.5,  # Entropy coefficient
        "C_sum": 0.7  # Cumulative context factor
    }

    @staticmethod
    def compress(data: Any, params: Dict[str, Any]) -> Dict[str, Any]:
        """Compress data using Möbius formula"""
        start_time = time.time()
        try:
            context = params.get("context", {})
            use_time_factor = params.get("useTimeFactor", True)
            use_energy_factor = params.get("useEnergyFactor", True)

            # Convert data to string for compression if it's not already
            if not isinstance(data, str):
                data_str = json.dumps(data)
            else:
                data_str = data

            # Calculate data size and entropy
            original_size = len(data_str)
            entropy = MobiusCompression.calculate_entropy(data_str)

            # Extract compression factors based on context
            factors = MobiusCompression.extract_context_factors(context)
            B = factors["B"]
            V = factors["V"]
            intent_factor = factors["I"]  # renamed from I to intent_factor to avoid ambiguity
            G = factors["G"]
            F = factors["F"]
            E = factors["E"] if use_energy_factor else 0.3
            t = factors["t"] if use_time_factor else 0.1
            C_sum = factors["C_sum"]

            # Calculate alignment as per Möbius formula
            alignment = (B + V * intent_factor) * math.exp(-t * E)

            # Numeric representation of the data (can be enhanced in future)
            value = MobiusCompression.get_numeric_representation(data_str)

            # Apply the Möbius Compression Formula
            entropy_factor = 1 - (entropy / math.log2(1 + V))
            numerator = value * B * intent_factor * entropy_factor * (G + F)
            denominator = E * t + C_sum * entropy + alignment

            # Guard against division by zero
            if abs(denominator) < 1e-10:
                denominator = 1e-10

            compressed_value = numerator / denominator

            # Create compressed representation
            compressed_package = {
                "algorithm": "mobius7d",
                "version": "1.0",
                "compressed": compressed_value,
                "value": value,
                "entropy": entropy,
                "compressionFactors": factors,
                "data": data_str  # Still include original data for this implementation
            }

            # Serialize to string for transmission
            compressed_json = json.dumps(compressed_package)
            compressed_data = compressed_json
            compressed_size = len(compressed_data)

            # Calculate compression ratio
            if original_size > 0:
                compression_ratio = 1.0 - (compressed_size / original_size)
            else:
                compression_ratio = 0.0

            # Update statistics
            MobiusCompression._stats["compressions"] += 1
            MobiusCompression._stats["totalOriginalSize"] += original_size
            MobiusCompression._stats["totalCompressedSize"] += compressed_size
            MobiusCompression._stats["lastRun"] = int(time.time() * 1000)
            
            # Return compression result with full context
            return {
                "success": True,
                "originalSize": original_size,
                "compressedSize": compressed_size,
                "compressionRatio": compression_ratio,
                "data": compressed_data,
                "metadata": {
                    "algorithm": "mobius7d",
                    "version": "1.0",
                    "B": B,
                    "V": V,
                    "I": intent_factor,
                    "G": G,
                    "F": F,
                    "E": E,
                    "t": t,
                    "C_sum": C_sum,
                    "entropy": entropy,
                    "value": value,
                    "alignment": alignment,
                    "useTimeFactor": use_time_factor,
                    "useEnergyFactor": use_energy_factor,
                },
                "contextVector": context,
                "timestamp": int(time.time() * 1000),
            }
        except Exception as e:
            MobiusCompression._stats["errors"] += 1
            return {
                "success": False,
                "error": f"Möbius compression failed: {str(e)}",
                "data": json.dumps(data),
                "timestamp": int(time.time() * 1000),
            }
        finally:
            MobiusCompression._stats["compressionTime"] += (time.time() - start_time)

    @staticmethod
    def decompress(compressed_data: Dict[str, Any]) -> Dict[str, Any]:
        """Decompress data using Möbius formula"""
        start_time = time.time()
        try:
            MobiusCompression._stats["decompressions"] += 1
            
            # Extract compressed data and metadata
            data = compressed_data.get("data", "")
            metadata = compressed_data.get("metadata", {})

            # Parse JSON
            decoded_json = json.loads(data)
            
            # If not using proper Möbius format, fall back to old method
            if "algorithm" not in decoded_json or decoded_json["algorithm"] != "mobius7d":
                # Legacy fallback
                try:
                    original_data = decoded_json["data"]
                    if isinstance(original_data, str):
                        # Try to parse as JSON if it looks like JSON
                        if original_data.strip().startswith(("{", "[")):
                            try:
                                original_data = json.loads(original_data)
                            except json.JSONDecodeError:
                                pass
                    return {
                        "success": True,
                        "data": original_data,
                        "metadata": metadata,
                        "timestamp": int(time.time() * 1000),
                    }
                except Exception:
                    # If any error, return the raw decoded JSON
                    return {
                        "success": True,
                        "data": decoded_json,
                        "metadata": metadata,
                        "timestamp": int(time.time() * 1000),
                    }
            
            # Extract compression variables
            compressed_value = decoded_json.get("compressed", 0)
            value = decoded_json.get("value", 0)
            entropy = decoded_json.get("entropy", 0)
            factors = decoded_json.get("compressionFactors", {})
            
            # If we have the original data in the package (for this implementation), use it directly
            if "data" in decoded_json:
                original_data = decoded_json["data"]
                if isinstance(original_data, str):
                    # Try to parse as JSON if it looks like JSON
                    if original_data.strip().startswith(("{", "[")):
                        try:
                            original_data = json.loads(original_data)
                        except json.JSONDecodeError:
                            pass
                return {
                    "success": True,
                    "data": original_data,
                    "metadata": metadata,
                    "timestamp": int(time.time() * 1000),
                }
            
            # Otherwise, perform an inverse of the Möbius formula (future implementation)
            # For now, this is a placeholder that would be filled in with actual inverse calculation
            # Inverse formula: value = compressed * (E * t + C_sum * entropy + alignment) / (B * I * (1 - entropy/log2(1 + V)) * (G + F))
            B = factors.get("B", MobiusCompression.DEFAULT_FACTORS["B"])
            V = factors.get("V", MobiusCompression.DEFAULT_FACTORS["V"])
            intent_factor = factors.get("I", MobiusCompression.DEFAULT_FACTORS["I"])
            G = factors.get("G", MobiusCompression.DEFAULT_FACTORS["G"])
            F = factors.get("F", MobiusCompression.DEFAULT_FACTORS["F"])
            E = factors.get("E", MobiusCompression.DEFAULT_FACTORS["E"])
            t = factors.get("t", 1.0)
            C_sum = factors.get("C_sum", MobiusCompression.DEFAULT_FACTORS["C_sum"])
            
            # Calculate alignment
            alignment = (B + V * intent_factor) * math.exp(-t * E)
            
            # Calculate inverse of Möbius formula
            entropy_factor = 1 - (entropy / math.log2(1 + V))
            denominator = B * intent_factor * entropy_factor * (G + F)
            numerator = E * t + C_sum * entropy + alignment
            
            # Guard against division by zero
            if abs(denominator) < 1e-10:
                denominator = 1e-10
                
            reconstructed_value = compressed_value * numerator / denominator
            
            # This is a placeholder - in a real implementation we would reconstruct the data
            reconstructed_data = {
                "decompressedFrom": "möbius",
                "reconstructedValue": reconstructed_value,
                "originalValue": value,
                "fidelity": max(0, min(1, reconstructed_value / (value + 1e-10))),
                "note": "This is a partial implementation that doesn't fully reconstruct the data"
            }
            
            return {
                "success": True,
                "data": reconstructed_data,
                "metadata": metadata,
                "timestamp": int(time.time() * 1000),
            }
        except Exception as e:
            MobiusCompression._stats["errors"] += 1
            return {
                "success": False,
                "error": f"Möbius decompression failed: {str(e)}",
                "timestamp": int(time.time() * 1000),
            }
        finally:
            MobiusCompression._stats["decompressionTime"] += (time.time() - start_time)
            MobiusCompression._stats["lastRun"] = int(time.time() * 1000)

    @staticmethod
    def get_statistics() -> Dict[str, Any]:
        """Get compression statistics"""
        stats = MobiusCompression._stats
        compressions = max(1, stats["compressions"])
        decompressions = max(1, stats["decompressions"])
        
        return {
            "compressionRatio": stats["totalCompressedSize"] / max(1, stats["totalOriginalSize"]),
            "averageCompressionTime": stats["compressionTime"] / compressions,
            "averageDecompressionTime": stats["decompressionTime"] / decompressions,
            "totalCompressed": stats["totalCompressedSize"],
            "totalDecompressed": stats["totalOriginalSize"],
            "compressions": stats["compressions"],
            "decompressions": stats["decompressions"],
            "errors": stats["errors"],
            "timestamp": int(time.time() * 1000),
        }
        
    @staticmethod
    def calculate_entropy(data: str) -> float:
        """
        Calculate Shannon entropy of the input data
        
        Args:
            data: Input string data
        
        Returns:
            Entropy value
        """
        # Count character frequencies
        frequencies = {}
        for char in data:
            frequencies[char] = frequencies.get(char, 0) + 1
        
        # Calculate entropy
        entropy = 0
        data_len = len(data)
        for count in frequencies.values():
            probability = count / data_len
            entropy -= probability * math.log2(probability)
        
        return entropy
    
    @staticmethod
    def extract_context_factors(context: Dict[str, Any]) -> Dict[str, float]:
        """
        Extract contextual factors based on the 7D context
        
        Args:
            context: The 7D context dictionary
            
        Returns:
            Dictionary of extracted compression factors
        """
        # Start with default factors
        factors = MobiusCompression.DEFAULT_FACTORS.copy()
        
        # Extract context-specific adjustments
        B = factors["B"]
        V = factors["V"]
        intent_factor = factors["I"]  # renamed to avoid ambiguity
        G = factors["G"]
        F = factors["F"]
        E = factors["E"]
        C_sum = factors["C_sum"]
        
        # WHO dimension influences base factor
        if context.get("who"):
            who_factor = 1.0
            if context["who"] == "MCPBridge":
                who_factor = 1.2
            elif "System" in str(context["who"]):
                who_factor = 1.1
            B *= who_factor
        
        # WHAT dimension influences variance
        if context.get("what"):
            what_factor = 1.0
            if "Compression" in str(context["what"]):
                what_factor = 1.3
            elif "Transform" in str(context["what"]):
                what_factor = 1.2
            V *= what_factor
        
        # WHY dimension influences intent factor
        if context.get("why"):
            why_factor = 1.0
            if "Optimize" in str(context["why"]):
                why_factor = 1.2
            elif "Protocol" in str(context["why"]):
                why_factor = 1.1
            intent_factor *= why_factor
            
        # WHEN dimension influences temporal gradient
        if context.get("when"):
            try:
                # If it's a timestamp, calculate relative to current time
                when = float(context["when"]) / 1000 if isinstance(context["when"], int) else time.time()
                # Calculate recency factor (more recent = higher G value)
                time_diff = max(0.1, min(3600, abs(time.time() - when)))
                G *= 1 / (1 + math.log(time_diff) / 10)
            except (ValueError, TypeError):
                pass
                
        # EXTENT dimension directly affects fidelity factor
        if context.get("extent"):
            try:
                extent = float(context["extent"])
                F = max(0.3, min(1.8, extent * 1.5))
            except (ValueError, TypeError):
                pass
                
        # Calculate time factor
        t = MobiusCompression.get_temporal_factor(context.get("when"))
        
        return {
            "B": B,
            "V": V,
            "I": intent_factor,
            "G": G,
            "F": F,
            "E": E,
            "t": t,
            "C_sum": C_sum
        }
    
    @staticmethod
    def get_temporal_factor(when) -> float:
        """
        Calculate temporal factor based on when dimension
        
        Args:
            when: Timestamp or temporal context
            
        Returns:
            Temporal factor value
        """
        if not when:
            return 1.0
            
        try:
            if isinstance(when, (int, float)):
                # Convert milliseconds to seconds if needed
                timestamp = when / 1000 if when > 1e10 else when
                # Calculate factor based on recency (more recent = lower t)
                now = time.time()
                time_diff = abs(now - timestamp)
                # Factor decreases as time difference increases (bounded)
                return max(0.1, min(2.0, 1.0 / (1 + math.log(1 + time_diff / 3600))))
            else:
                # For non-timestamp when values, use a default
                return 1.0
        except (ValueError, TypeError):
            return 1.0
    
    @staticmethod
    def get_numeric_representation(data: str) -> float:
        """
        Convert string data to numeric representation for compression
        
        Args:
            data: String data to be represented numerically
            
        Returns:
            Numeric representation
        """
        # Use string length as base value
        value = len(data)
        
        # Incorporate character code information with diminishing influence
        for i, char in enumerate(data[:100]):  # Limit to first 100 chars for performance
            value += ord(char) / (i + 1)
            
        # Add entropy contribution
        entropy = MobiusCompression.calculate_entropy(data[:1000])  # Sample entropy
        value += entropy * 100
            
        return value
