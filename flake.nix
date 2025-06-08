{  
  description = "Dupee developer enviroment";  
  
  inputs = {  
    nixpkgs.url = "nixpkgs";  
  };  
  
  outputs = { self, nixpkgs }:   
  let   
      # Systems supported  
      allSystems = [  
        "x86_64-linux" # 64-bit Intel/AMD Linux  
        "aarch64-linux" # 64-bit ARM Linux  
        "x86_64-darwin" # 64-bit Intel macOS  
        "aarch64-darwin" # 64-bit ARM macOS  
      ];  
  
      # Helper to provide system-specific attributes  
      forAllSystems = f: nixpkgs.lib.genAttrs allSystems (system: f {  
        pkgs = import nixpkgs { inherit system; };  
      });  
  in {  
    # Development env package required.  
    devShells = forAllSystems ({ pkgs }: {  
        default = pkgs.mkShell {  
          # The Nix packages provided in the environment  
          packages = with pkgs; [  
            go # Go 1.22  
            gotools # Go tools like goimports, godoc, and others  
          ];  
        };  
    });  
  };  
}        
