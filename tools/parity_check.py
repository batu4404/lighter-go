#!/usr/bin/env python3
"""
Cross-language signing parity check: lighter-go (this fork) vs lighter-python.

Builds this fork's signer shared library, loads it side-by-side with the
shared library bundled in ../lighter-python (the reference, compiled from the
upstream signer for a tagged release), signs an identical set of transactions
through both via ctypes, and asserts the signed output is byte-identical.

Two outputs are inherently non-deterministic and are excluded from the
comparison:
  * Sig          -- Schnorr signatures use a fresh random nonce each call
  * ExpiredAt    -- defaults to time.Now()+10min, set internally (not a C arg),
                    so it drifts a few ms between the two calls and feeds the hash
Everything else -- the tx type, every semantic field, messageToSign and any
validation error -- must match exactly. When the two ExpiredAt values happen to
land in the same millisecond, the full txHash is asserted identical too.

Usage:
    python3 tools/parity_check.py            # builds the fork dylib, then compares
    REF=/path/to/lighter-signer.so python3 tools/parity_check.py
"""

import ctypes
import json
import os
import platform
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
GO_ROOT = os.path.dirname(HERE)
PY_SIGNERS = os.path.normpath(os.path.join(GO_ROOT, "..", "lighter-python", "lighter", "signers"))


def platform_signer_name():
    system = platform.system()
    machine = platform.machine().lower()
    is_x64 = machine in ("amd64", "x86_64")
    is_arm = machine == "arm64"
    if system == "Darwin" and is_arm:
        return "lighter-signer-darwin-arm64.dylib", "dylib"
    if system == "Linux" and is_x64:
        return "lighter-signer-linux-amd64.so", "so"
    if system == "Linux" and is_arm:
        return "lighter-signer-linux-arm64.so", "so"
    if system == "Windows" and is_x64:
        return "lighter-signer-windows-amd64.dll", "dll"
    raise SystemExit(f"unsupported platform: {system}/{machine}")


def build_fork_lib(ext):
    out = os.path.join(HERE, f"_parity_fork_signer.{ext}")
    env = dict(os.environ, CGO_ENABLED="1")
    print(f"building fork signer -> {out}")
    subprocess.run(
        ["go", "build", "-buildmode=c-shared", "-trimpath", "-o", out, "./sharedlib/main.go"],
        cwd=GO_ROOT, env=env, check=True,
    )
    return out


# ---- ctypes structures (mirror lighter-python v1.1.1 signer_client.py) ----
class ApiKeyResponse(ctypes.Structure):
    _fields_ = [("privateKey", ctypes.c_char_p), ("publicKey", ctypes.c_char_p), ("err", ctypes.c_char_p)]


class CreateOrderTxReq(ctypes.Structure):
    _fields_ = [
        ("MarketIndex", ctypes.c_int), ("ClientOrderIndex", ctypes.c_longlong),
        ("BaseAmount", ctypes.c_longlong), ("Price", ctypes.c_uint32),
        ("IsAsk", ctypes.c_uint8), ("Type", ctypes.c_uint8), ("TimeInForce", ctypes.c_uint8),
        ("ReduceOnly", ctypes.c_uint8), ("TriggerPrice", ctypes.c_uint32), ("OrderExpiry", ctypes.c_longlong)]


class SignedTxResponse(ctypes.Structure):
    _fields_ = [("txType", ctypes.c_uint8), ("txInfo", ctypes.c_char_p), ("txHash", ctypes.c_char_p),
                ("messageToSign", ctypes.c_char_p), ("err", ctypes.c_char_p)]


def setup(path):
    s = ctypes.CDLL(path)
    s.GenerateAPIKey.argtypes = []
    s.GenerateAPIKey.restype = ApiKeyResponse
    s.CreateClient.argtypes = [ctypes.c_char_p, ctypes.c_char_p, ctypes.c_int, ctypes.c_int, ctypes.c_longlong]
    s.CreateClient.restype = ctypes.c_void_p
    sig = SignedTxResponse
    s.SignCreateOrder.argtypes = [ctypes.c_int, ctypes.c_longlong, ctypes.c_longlong, ctypes.c_int, ctypes.c_int, ctypes.c_int, ctypes.c_int, ctypes.c_int, ctypes.c_int, ctypes.c_longlong, ctypes.c_longlong, ctypes.c_int, ctypes.c_int, ctypes.c_uint8, ctypes.c_uint8, ctypes.c_uint8, ctypes.c_longlong, ctypes.c_int, ctypes.c_longlong]; s.SignCreateOrder.restype = sig
    s.SignCancelOrder.argtypes = [ctypes.c_int, ctypes.c_longlong, ctypes.c_uint8, ctypes.c_longlong, ctypes.c_int, ctypes.c_longlong]; s.SignCancelOrder.restype = sig
    s.SignCancelAllOrders.argtypes = [ctypes.c_int, ctypes.c_longlong, ctypes.c_int, ctypes.c_uint8, ctypes.c_longlong, ctypes.c_int, ctypes.c_longlong]; s.SignCancelAllOrders.restype = sig
    s.SignWithdraw.argtypes = [ctypes.c_int, ctypes.c_int, ctypes.c_longlong, ctypes.c_uint8, ctypes.c_longlong, ctypes.c_int, ctypes.c_longlong]; s.SignWithdraw.restype = sig
    s.SignCreateSubAccount.argtypes = [ctypes.c_uint8, ctypes.c_longlong, ctypes.c_int, ctypes.c_longlong]; s.SignCreateSubAccount.restype = sig
    s.SignModifyOrder.argtypes = [ctypes.c_int, ctypes.c_longlong, ctypes.c_longlong, ctypes.c_longlong, ctypes.c_longlong, ctypes.c_longlong, ctypes.c_int, ctypes.c_int, ctypes.c_uint8, ctypes.c_uint8, ctypes.c_uint8, ctypes.c_longlong, ctypes.c_int, ctypes.c_longlong]; s.SignModifyOrder.restype = sig
    s.SignTransfer.argtypes = [ctypes.c_longlong, ctypes.c_int16, ctypes.c_int8, ctypes.c_int8, ctypes.c_longlong, ctypes.c_longlong, ctypes.c_char_p, ctypes.c_uint8, ctypes.c_longlong, ctypes.c_int, ctypes.c_longlong]; s.SignTransfer.restype = sig
    s.SignCreatePublicPool.argtypes = [ctypes.c_longlong, ctypes.c_int, ctypes.c_longlong, ctypes.c_uint8, ctypes.c_longlong, ctypes.c_int, ctypes.c_longlong]; s.SignCreatePublicPool.restype = sig
    s.SignUpdatePublicPool.argtypes = [ctypes.c_longlong, ctypes.c_int, ctypes.c_longlong, ctypes.c_int, ctypes.c_uint8, ctypes.c_longlong, ctypes.c_int, ctypes.c_longlong]; s.SignUpdatePublicPool.restype = sig
    s.SignMintShares.argtypes = [ctypes.c_longlong, ctypes.c_longlong, ctypes.c_uint8, ctypes.c_longlong, ctypes.c_int, ctypes.c_longlong]; s.SignMintShares.restype = sig
    s.SignBurnShares.argtypes = [ctypes.c_longlong, ctypes.c_longlong, ctypes.c_uint8, ctypes.c_longlong, ctypes.c_int, ctypes.c_longlong]; s.SignBurnShares.restype = sig
    s.SignUpdateLeverage.argtypes = [ctypes.c_int, ctypes.c_int, ctypes.c_int, ctypes.c_uint8, ctypes.c_longlong, ctypes.c_int, ctypes.c_longlong]; s.SignUpdateLeverage.restype = sig
    s.SignUpdateMargin.argtypes = [ctypes.c_int, ctypes.c_longlong, ctypes.c_int, ctypes.c_uint8, ctypes.c_longlong, ctypes.c_int, ctypes.c_longlong]; s.SignUpdateMargin.restype = sig
    s.SignStakeAssets.argtypes = [ctypes.c_longlong, ctypes.c_longlong, ctypes.c_uint8, ctypes.c_longlong, ctypes.c_int, ctypes.c_longlong]; s.SignStakeAssets.restype = sig
    s.SignUnstakeAssets.argtypes = [ctypes.c_longlong, ctypes.c_longlong, ctypes.c_uint8, ctypes.c_longlong, ctypes.c_int, ctypes.c_longlong]; s.SignUnstakeAssets.restype = sig
    s.SignApproveIntegrator.argtypes = [ctypes.c_longlong, ctypes.c_uint32, ctypes.c_uint32, ctypes.c_uint32, ctypes.c_uint32, ctypes.c_longlong, ctypes.c_uint8, ctypes.c_longlong, ctypes.c_int, ctypes.c_longlong]; s.SignApproveIntegrator.restype = sig
    s.SignUpdateAccountConfig.argtypes = [ctypes.c_uint8, ctypes.c_uint8, ctypes.c_longlong, ctypes.c_int, ctypes.c_longlong]; s.SignUpdateAccountConfig.restype = sig
    s.SignUpdateAccountAssetConfig.argtypes = [ctypes.c_int16, ctypes.c_uint8, ctypes.c_uint8, ctypes.c_longlong, ctypes.c_int, ctypes.c_longlong]; s.SignUpdateAccountAssetConfig.restype = sig
    return s


def res(r):
    if r.err:
        return {"err": r.err.decode()}, None, None
    info = json.loads(r.txInfo.decode()) if r.txInfo else {}
    expired = info.pop("ExpiredAt", None)
    info.pop("Sig", None)
    msg = r.messageToSign.decode() if r.messageToSign else None
    txhash = r.txHash.decode() if r.txHash else None
    return (r.txType, info, msg), expired, txhash


def main():
    ref_name, ext = platform_signer_name()
    ref = os.environ.get("REF", os.path.join(PY_SIGNERS, ref_name))
    if not os.path.exists(ref):
        raise SystemExit(f"reference signer not found: {ref}")
    fork = build_fork_lib(ext)

    print(f"ref  = {ref}")
    print(f"fork = {fork}\n")
    R = setup(ref)
    F = setup(fork)

    key = R.GenerateAPIKey()
    PRIV = key.privateKey
    CHAIN, AK, ACC = 304, 0, 1
    for name, s in (("ref", R), ("fork", F)):
        err = s.CreateClient(b"", PRIV, CHAIN, AK, ACC)
        if err:
            raise SystemExit(f"CreateClient {name} failed: {ctypes.cast(err, ctypes.c_char_p).value}")

    NONCE = 42
    POOL = 140737488355328  # (1 << 47), the staking / public-pool index offset
    MEMO = b"0" * 64        # 32-byte memo as 64 hex chars
    cases = [
        # CreateOrder args (v1.1.1): marketIndex, clientOrderIndex, baseAmount, price,
        # isAsk, orderType, timeInForce, reduceOnly, triggerPrice, orderExpiry,
        # integratorAccountIndex, integratorTakerFee, integratorMakerFee,
        # selfTradeBehaviorMode, selfTradeEqualityMode, skipNonce, nonce, apiKeyIndex, accountIndex
        ("CreateOrder",              lambda s: s.SignCreateOrder(1, 0, 1000, 350000, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, NONCE, AK, ACC)),
        ("CreateOrder+skip",         lambda s: s.SignCreateOrder(1, 0, 1000, 350000, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, NONCE, AK, ACC)),
        # self-trade modes: behavior=1 (expire taker), equality=1 (master account index)
        ("CreateOrder+selftrade",    lambda s: s.SignCreateOrder(1, 0, 1000, 350000, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, NONCE, AK, ACC)),
        ("CancelOrder",              lambda s: s.SignCancelOrder(1, 7, 0, NONCE, AK, ACC)),
        ("CancelOrder+skip",         lambda s: s.SignCancelOrder(1, 7, 1, NONCE, AK, ACC)),
        # CancelAllOrders args (v1.1.1): timeInForce, time, cancelAllMarketIndex, skipNonce, nonce, apiKeyIndex, accountIndex
        ("CancelAllOrders",          lambda s: s.SignCancelAllOrders(0, 123, 255, 0, NONCE, AK, ACC)),
        ("CancelAllOrders+market",   lambda s: s.SignCancelAllOrders(0, 0, 1, 0, NONCE, AK, ACC)),
        ("Withdraw",                 lambda s: s.SignWithdraw(3, 0, 5000000, 0, NONCE, AK, ACC)),
        ("CreateSubAccount",         lambda s: s.SignCreateSubAccount(0, NONCE, AK, ACC)),
        # ModifyOrder args (v1.1.1): marketIndex, index, baseAmount, price, triggerPrice,
        # integratorAccountIndex, integratorTakerFee, integratorMakerFee,
        # selfTradeBehaviorMode, selfTradeEqualityMode, skipNonce, nonce, apiKeyIndex, accountIndex
        ("ModifyOrder",              lambda s: s.SignModifyOrder(1, 7, 2000, 360000, 0, 0, 0, 0, 0, 0, 0, NONCE, AK, ACC)),
        ("ModifyOrder+selftrade",    lambda s: s.SignModifyOrder(1, 7, 2000, 360000, 0, 0, 0, 0, 1, 0, 0, NONCE, AK, ACC)),
        ("Transfer",                 lambda s: s.SignTransfer(2, 3, 0, 0, 1000000, 100, MEMO, 0, NONCE, AK, ACC)),
        ("CreatePublicPool",         lambda s: s.SignCreatePublicPool(100, 1000, 50, 0, NONCE, AK, ACC)),
        ("UpdatePublicPool",         lambda s: s.SignUpdatePublicPool(POOL, 1, 100, 50, 0, NONCE, AK, ACC)),
        ("MintShares",               lambda s: s.SignMintShares(POOL, 500, 0, NONCE, AK, ACC)),
        ("BurnShares",               lambda s: s.SignBurnShares(POOL, 500, 0, NONCE, AK, ACC)),
        ("UpdateLeverage",           lambda s: s.SignUpdateLeverage(1, 2000, 0, 0, NONCE, AK, ACC)),
        ("UpdateMargin",             lambda s: s.SignUpdateMargin(1, 1000000, 0, 0, NONCE, AK, ACC)),
        ("StakeAssets",              lambda s: s.SignStakeAssets(POOL, 500, 0, NONCE, AK, ACC)),
        ("StakeAssets+skip",         lambda s: s.SignStakeAssets(POOL, 500, 1, NONCE, AK, ACC)),
        ("UnstakeAssets",            lambda s: s.SignUnstakeAssets(POOL, 500, 0, NONCE, AK, ACC)),
        ("ApproveIntegrator",        lambda s: s.SignApproveIntegrator(5, 10, 20, 30, 40, 999, 0, NONCE, AK, ACC)),
        # zero-fee approval with non-zero expiry (allowed since #77)
        ("ApproveIntegrator+zerofee", lambda s: s.SignApproveIntegrator(5, 0, 0, 0, 0, 999, 0, NONCE, AK, ACC)),
        ("UpdateAccountConfig",      lambda s: s.SignUpdateAccountConfig(1, 0, NONCE, AK, ACC)),
        ("UpdateAccountAssetConfig", lambda s: s.SignUpdateAccountAssetConfig(3, 1, 0, NONCE, AK, ACC)),
    ]

    npass = nfail = hash_proofs = 0
    for name, fn in cases:
        a, ea, ha = res(fn(R))
        b, eb, hb = res(fn(F))
        ok = (a == b)
        note = ""
        if ok and ea is not None and ea == eb:
            if ha == hb:
                hash_proofs += 1
                note = "  [txHash identical]"
            else:
                ok = False
                note = f"\n  txHash DIFFERS ref={ha} fork={hb}"
        print(("PASS" if ok else "FAIL"), name, note if ok else f"\n  ref ={a}\n  fork={b}")
        npass += ok
        nfail += (not ok)

    os.remove(fork)
    header = os.path.splitext(fork)[0] + ".h"
    if os.path.exists(header):
        os.remove(header)

    print(f"\n{npass} identical / {nfail} mismatched out of {len(cases)}")
    print(f"{hash_proofs} also had byte-identical txHash (ExpiredAt coincided)")
    return 1 if nfail else 0


if __name__ == "__main__":
    sys.exit(main())
