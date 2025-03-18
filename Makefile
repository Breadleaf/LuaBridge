# NOTE
# I am assuming you have gcc, wget, zip all installed before you run this

LUA_VERSION := 5.4.7
LUA_RELEASE := https://github.com/lua/lua/archive/v$(LUA_VERSION).zip
LUA_TEMP_DIR := lua-$(LUA_VERSION)
LUA_ZIP := $(LUA_TEMP_DIR).zip

# Dont change these
LUA_INSTALL_DIR := ./lua
LUA_WRAPPER_FILE := wrapper.go
TEMP_STORAGE := .storage_$(LUA_WRAPPER_FILE)

.PHONY: all clean_lua clean store_patch retrieve_patch install_helper

all: $(LUA_INSTALL_DIR)

store_patch:
	@if [ -f $(LUA_INSTALL_DIR)/$(LUA_WRAPPER_FILE) ]; then \
		echo "Storing $(LUA_WRAPPER_FILE) to $(TEMP_STORAGE)"; \
		cp $(LUA_INSTALL_DIR)/$(LUA_WRAPPER_FILE) $(TEMP_STORAGE); \
	else \
		echo "$(LUA_WRAPPER_FILE) not found in $(LUA_INSTALL_DIR), nothing to store."; \
	fi

retrieve_patch:
	@if [ -f $(TEMP_STORAGE) ]; then \
		echo "Restoring $(LUA_WRAPPER_FILE) from $(TEMP_STORAGE) to $(LUA_INSTALL_DIR)"; \
		cp $(TEMP_STORAGE) $(LUA_INSTALL_DIR)/$(LUA_WRAPPER_FILE); \
	else \
		echo "No stored $(LUA_WRAPPER_FILE) found in $(TEMP_STORAGE), skipping restore."; \
	fi
	@if [ ! -f $(LUA_INSTALL_DIR)/$(LUA_WRAPPER_FILE) ]; then \
		echo "Error: $(LUA_WRAPPER_FILE) is missing in $(LUA_INSTALL_DIR). Please reinstall the project from git."; \
		exit 1; \
	fi

install_helper: $(LUA_ZIP)
	mkdir -p $(LUA_INSTALL_DIR)
	unzip $(LUA_ZIP)
	mv $(LUA_TEMP_DIR)/* $(LUA_INSTALL_DIR)/
	# Patch the lua install
	rm -rf \
		$(LUA_INSTALL_DIR)/lua.c \
		$(LUA_INSTALL_DIR)/ltests.* \
		$(LUA_INSTALL_DIR)/onelua.c \
		$(LUA_INSTALL_DIR)/makefile \
		$(LUA_INSTALL_DIR)/manual \
		$(LUA_INSTALL_DIR)/all \
		$(LUA_INSTALL_DIR)/testes/ \
		$(LUA_INSTALL_DIR)/README.md
	rm -rf $(LUA_TEMP_DIR)
	rm -f $(LUA_ZIP)

$(LUA_INSTALL_DIR): store_patch install_helper retrieve_patch 

$(LUA_ZIP):
	wget $(LUA_RELEASE) -O $(LUA_ZIP)

clean_lua:
	@if [ -f $(LUA_INSTALL_DIR)/$(LUA_WRAPPER_FILE) ]; then \
		echo "Preserving $(LUA_WRAPPER_FILE) before cleaning the Lua install."; \
		cp $(LUA_INSTALL_DIR)/$(LUA_WRAPPER_FILE) $(TEMP_STORAGE); \
	fi
	rm -rf $(LUA_INSTALL_DIR)

clean: clean_lua
	rm -rf $(LUA_ZIP)
