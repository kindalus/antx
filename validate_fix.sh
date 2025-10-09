#!/bin/bash

# Validation script for Current Node Persistence Fix
# This script helps validate that the currentNode is saved correctly after navigation

set -e

echo "🔧 Antbox CLI Current Node Persistence Fix Validation"
echo "====================================================="
echo

# Check if antx binary exists
if [[ ! -f "./antx" ]]; then
    echo "❌ antx binary not found. Please run 'go build' first."
    exit 1
fi

# Create a backup of existing config if it exists
CONFIG_FILE="$HOME/.antx"
BACKUP_FILE="$HOME/.antx.backup.$(date +%s)"

if [[ -f "$CONFIG_FILE" ]]; then
    echo "📁 Backing up existing config to: $BACKUP_FILE"
    cp "$CONFIG_FILE" "$BACKUP_FILE"
    echo
fi

echo "🧪 Test Setup"
echo "============"
echo "This script will help you manually validate that the current node"
echo "persistence fix is working correctly."
echo
echo "The fix ensures that navigation commands (cd, etc.) save the CORRECT"
echo "current location to ~/.antx, not the previous location."
echo
echo "To run this test, you'll need:"
echo "1. A running Antbox server"
echo "2. Valid credentials (API key, JWT, or root password)"
echo "3. At least one folder to navigate to"
echo

# Function to show config file content
show_config() {
    if [[ -f "$CONFIG_FILE" ]]; then
        echo "Current ~/.antx content:"
        echo "------------------------"
        local line_num=0
        while IFS= read -r line; do
            line_num=$((line_num + 1))
            if [[ $line_num -eq 1 ]]; then
                echo "Current Node: $line"
            elif [[ $line_num -eq 2 ]]; then
                echo "Separator: (blank line)"
            elif [[ -n "$line" ]]; then
                echo "History: $line"
            fi
        done < "$CONFIG_FILE"
        echo "------------------------"
    else
        echo "No ~/.antx file found"
    fi
    echo
}

# Function to wait for user input
wait_for_user() {
    read -p "Press Enter to continue..."
    echo
}

echo "📋 Validation Steps"
echo "=================="
echo "Follow these steps to validate the fix:"
echo

echo "Step 1: Initial State"
echo "--------------------"
echo "First, let's see the initial state of your config file:"
show_config

echo "Step 2: Start CLI"
echo "----------------"
echo "Start the CLI with your server and credentials. For example:"
echo "  ./antx http://localhost:8080 --root=yourpassword"
echo "OR"
echo "  ./antx http://localhost:8080 --api-key=yourkey"
echo
echo "After the CLI starts, you should see a prompt like:"
echo "  antx (root)> "
echo

wait_for_user

echo "Step 3: Check Initial Location"
echo "-----------------------------"
echo "In the CLI, run these commands to see your starting location:"
echo "  pwd"
echo "  status"
echo
echo "Note down what your current location is."
echo

wait_for_user

echo "Step 4: Navigate to a Folder"
echo "----------------------------"
echo "List available folders and navigate to one:"
echo "  ls"
echo "  cd <some-folder-uuid>"
echo
echo "Replace <some-folder-uuid> with an actual folder UUID from the ls output."
echo

wait_for_user

echo "Step 5: Check Config After Navigation"
echo "------------------------------------"
echo "Exit the CLI (type 'exit') and then check the config file:"
show_config

echo "❗ VALIDATION POINT #1:"
echo "The 'Current Node' line should show the UUID of the folder you navigated to,"
echo "NOT the previous location (root). If it shows the folder UUID, the fix is working!"
echo

wait_for_user

echo "Step 6: Navigate Again"
echo "---------------------"
echo "Restart the CLI and navigate to a different folder (or back):"
echo "1. Start CLI again: ./antx http://localhost:8080 --credentials"
echo "2. Verify you're restored to the correct location: pwd"
echo "3. Navigate somewhere else: cd <another-uuid> or cd .."
echo "4. Exit CLI: exit"
echo

wait_for_user

echo "Step 7: Final Validation"
echo "-----------------------"
show_config

echo "❗ VALIDATION POINT #2:"
echo "The 'Current Node' should show your latest location, not the previous one."
echo

echo "🏁 Test Results"
echo "=============="
echo
read -p "Did the config file show the CORRECT current location after each navigation? (y/n): " result

if [[ "$result" =~ ^[Yy] ]]; then
    echo
    echo "🎉 SUCCESS! The current node persistence fix is working correctly."
    echo
    echo "✅ Navigation commands now save the correct current location"
    echo "✅ CLI will restore to the right place on restart"
    echo "✅ No more 'one step behind' behavior"
else
    echo
    echo "❌ ISSUE DETECTED: The fix may not be working properly."
    echo
    echo "Debugging steps:"
    echo "1. Verify you built the latest code: go build"
    echo "2. Check that prompt.go was modified correctly"
    echo "3. Look for any error messages during navigation"
    echo "4. Try the manual test again with different folders"
fi

echo

# Restore backup if user wants
if [[ -f "$BACKUP_FILE" ]]; then
    echo "💾 Config Backup"
    echo "==============="
    read -p "Do you want to restore your original config from backup? (y/n): " restore
    if [[ "$restore" =~ ^[Yy] ]]; then
        cp "$BACKUP_FILE" "$CONFIG_FILE"
        echo "✅ Original config restored from: $BACKUP_FILE"
    else
        echo "📁 Backup preserved at: $BACKUP_FILE"
    fi
    echo
fi

echo "🔍 Additional Debugging"
echo "======================"
echo "If you need to debug further, you can:"
echo "1. Watch the config file in real time: watch -n 1 cat ~/.antx"
echo "2. Run CLI with verbose output: ./antx --verbose ..."
echo "3. Check the exact timing by adding debug prints"
echo

echo "✅ Validation script complete!"
