# ============================================================================
# TEMPORAL NEURON - SIMPLE TESTING MAKEFILE
# ============================================================================

# Configuration
GO ?= go
TIMEOUT ?= 30s

# Colors
GREEN := \033[0;32m
YELLOW := \033[1;33m
CYAN := \033[0;36m
NC := \033[0m

.PHONY: help test quick full clean neuron synapse glial

# Default target
help:
	@echo "$(CYAN)Temporal Neuron Testing$(NC)"
	@echo "======================="
	@echo ""
	@echo "$(YELLOW)make quick$(NC)   - Fast shakedown test (< 30s)"
	@echo "$(YELLOW)make test$(NC)    - Run all package tests"
	@echo "$(YELLOW)make full$(NC)    - Everything including slow tests"
	@echo ""
	@echo "$(YELLOW)make neuron$(NC)  - Test neuron package"
	@echo "$(YELLOW)make synapse$(NC) - Test synapse package" 
	@echo "$(YELLOW)make glial$(NC)   - Test glial package"
	@echo ""
	@echo "$(YELLOW)make clean$(NC)   - Clean test cache"

# Quick shakedown - run this after any change
quick:
	@echo "$(CYAN)🚀 Quick Shakedown Test$(NC)"
	@$(GO) test -short -timeout=$(TIMEOUT) -v ./neuron -run "TestNeuronCreation|TestReceiveMethod|TestThresholdBasedFiring"
	@$(GO) test -short -timeout=$(TIMEOUT) -v ./synapse -run "TestBasicSynapseCreation|TestSynapseTransmission"
	@$(GO) test -short -timeout=$(TIMEOUT) -v ./glial -run "TestBasicProcessingMonitor"
	@echo "$(GREEN)✅ Quick tests passed!$(NC)"

# Standard test - all working packages
test:
	@echo "$(CYAN)🧪 Running All Package Tests$(NC)"
	@$(GO) test -timeout=60s -v ./neuron ./synapse ./glial
	@echo "$(GREEN)✅ All tests passed!$(NC)"

# Full test suite including slow tests
full:
	@echo "$(CYAN)🎯 Full Test Suite$(NC)"
	@$(GO) test -timeout=120s -v ./...
	@echo "$(GREEN)✅ Full suite completed!$(NC)"

# Individual package tests
neuron:
	@echo "$(CYAN)🧠 Testing Neuron Package$(NC)"
	@$(GO) test -timeout=60s -v ./neuron

synapse:
	@echo "$(CYAN)🔗 Testing Synapse Package$(NC)"
	@$(GO) test -timeout=60s -v ./synapse

glial:
	@echo "$(CYAN)🌟 Testing Glial Package$(NC)"
	@$(GO) test -timeout=60s -v ./glial

# Clean test cache
clean:
	@echo "$(CYAN)🧹 Cleaning test cache$(NC)"
	@$(GO) clean -testcache
	@echo "$(GREEN)✅ Clean completed!$(NC)"

# Add new packages as you develop them
# Example: 
# extracellular:
# 	@echo "$(CYAN)🌐 Testing Extracellular Package$(NC)"
# 	@$(GO) test -timeout=60s -v ./extracellular