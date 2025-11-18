{ pkgs ? import <nixpkgs> {
    config = {
      allowUnfree = true;
    };
  }
}:

pkgs.mkShell {
  name = "pytorch-cpu-env";

  buildInputs = with pkgs; [
    python3
    python3Packages.pip
    python3Packages.virtualenv
    stdenv.cc.cc.lib
    zlib
    go-migrate
  ];

  shellHook = ''
    export LD_LIBRARY_PATH=${pkgs.stdenv.cc.cc.lib}/lib:${pkgs.zlib}/lib:$LD_LIBRARY_PATH
    echo "Python environment ready!"
    echo "Python version: $(python --version)"
    echo "Pip version: $(pip --version)"
    echo ""
    echo "To install CPU-only PyTorch and other packages, run:"
    echo "  pip install torch torchvision numpy pandas pillow timm --extra-index-url https://download.pytorch.org/whl/cpu"
    echo ""
    echo "Or create a virtual environment first:"
    echo "  python -m venv venv_cpu"
    echo "  source venv_cpu/bin/activate"
    echo "  pip install torch torchvision numpy pandas pillow timm --extra-index-url https://download.pytorch.org/whl/cpu"
  '';
}
