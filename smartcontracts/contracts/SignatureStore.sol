// SPDX-License-Identifier: MIT

pragma solidity ^0.8.0;

contract SignatureStore {
    mapping(string => string) public signatures;

    function saveSignature(string memory id, string memory signature) public {
        signatures[id] = signature;
    }

    function getSignature(string memory id) public view returns (string memory) {
        return signatures[id];
    }
}